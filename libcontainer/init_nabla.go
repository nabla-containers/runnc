// Copyright 2014 Docker, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package libcontainer

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/nabla-containers/runnc/libcontainer/configs"
	ll "github.com/nabla-containers/runnc/llif"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/vishvananda/netns"
)

type initConfig struct {
	Id         string          `json:"id"`
	BundlePath string          `json:"bundlepath"`
	Root       string          `json:"root"`
	Args       []string        `json:"args"`
	Cwd        string          `json:"cwd"`
	Env        []string        `json:"env"`
	TapName    string          `json:"tap"`
	NetnsPath  string          `json:"netnspath"`
	Hooks      *spec.Hooks     `json:"hooks"`
	Memory     int64           `json:"mem"`
	Mounts     []spec.Mount    `json:"Mounts"`
	Config     *configs.Config `json:"config"`

	FsState      ll.LLState `json:"fsstate"`
	NetworkState ll.LLState `json:"netstate"`
	ExecState    ll.LLState `json:execstate"`
}

func initNabla(llcHandler ll.RunllcHandler) error {
	var (
		pipefd, rootfd int
		envInitPipe    = os.Getenv("_LIBCONTAINER_INITPIPE")
		envStateDir    = os.Getenv("_LIBCONTAINER_STATEDIR")
	)

	// Get the INITPIPE.
	pipefd, err := strconv.Atoi(envInitPipe)
	if err != nil {
		return fmt.Errorf("unable to convert _LIBCONTAINER_INITPIPE=%s to int: %s", envInitPipe, err)
	}

	pipe := os.NewFile(uintptr(pipefd), "pipe")
	defer pipe.Close()

	var config *initConfig
	if err := json.NewDecoder(pipe).Decode(&config); err != nil {
		return err
	}

	// Only init processes have STATEDIR.
	if rootfd, err = strconv.Atoi(envStateDir); err != nil {
		return fmt.Errorf("unable to convert _LIBCONTAINER_STATEDIR=%s to int: %s", envStateDir, err)
	}

	// clear the current process's environment to clean any libcontainer
	// specific env vars.
	os.Clearenv()

	// LLC Fs Handle
	fsInput := &ll.FsRunInput{}
	fsInput.ContainerRoot = config.Root
	fsInput.Config = config.Config
	fsInput.ContainerId = config.Id
	fsInput.FsState = &config.FsState
	fsInput.NetworkState = &config.NetworkState
	fsInput.ExecState = &config.ExecState

	// TODO(runllc): Propagate and store LLstates
	fsState, err := llcHandler.FsH.FsRunFunc(fsInput)
	if err != nil {
		return fmt.Errorf("Error running llc Fs handler: %v", err)
	}

	// Go into network namespace for temporary hack for CNI plugin using veth pairs
	// K8s case
	if config.NetnsPath != "" {
		nsh, err := netns.GetFromPath(config.NetnsPath)
		if err != nil {
			return newSystemErrorWithCause(err, "unable to get netns handle")
		}

		if err := netns.Set(nsh); err != nil {
			return newSystemErrorWithCause(err, "unable to get set netns")
		}
	} else {
		// Docker case for docker cli
		// TODO: case on specific --docker-cli flag
		nsh, err := netns.New()
		if err != nil {
			return newSystemErrorWithCause(err, "unable to create netns handle")
		}

		if err := netns.Set(nsh); err != nil {
			return newSystemErrorWithCause(err, "unable to get set netns")
		}
	}
	if config.Hooks != nil {
		for _, hook := range config.Hooks.Prestart {
			if err := runHook(hook, config.Id, config.BundlePath); err != nil {
				return newSystemErrorWithCause(err, "unable to run prestart hook")
			}
		}
	}

	networkInput := &ll.NetworkRunInput{}
	networkInput.ContainerRoot = config.Root
	networkInput.Config = config.Config
	networkInput.ContainerId = config.Id
	networkInput.FsState = fsState
	networkInput.NetworkState = &config.NetworkState
	networkInput.ExecState = &config.ExecState

	// TODO(runllc): Propagate and store LLstates
	networkState, err := llcHandler.NetworkH.NetworkRunFunc(networkInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running llc Network handler: %v", err)
		return fmt.Errorf("Error running llc Network handler: %v", err)
	}

	// wait for the fifo to be opened on the other side before
	// exec'ing the users process.
	fd, err := syscall.Openat(rootfd, execFifoFilename, os.O_WRONLY|syscall.O_CLOEXEC, 0)
	if err != nil {
		return newSystemErrorWithCause(err, "openat exec fifo")
	}
	if _, err := syscall.Write(fd, []byte("0")); err != nil {
		return newSystemErrorWithCause(err, "write 0 exec fifo")
	}
	syscall.Close(fd)
	syscall.Close(rootfd)

	// Check if it is a pause container, if it is, just pause
	if len(config.Args) == 1 && config.Args[0] == pauseNablaName {
		select {}
	}

	// LLC Exec Handle
	execInput := &ll.ExecRunInput{}
	execInput.ContainerRoot = config.Root
	execInput.Config = config.Config
	execInput.ContainerId = config.Id
	execInput.FsState = fsState
	execInput.NetworkState = networkState
	execInput.ExecState = &config.ExecState

	// Should not return if successful
	return llcHandler.ExecH.ExecRunFunc(execInput)
}
