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
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/nabla-containers/runnc/libcontainer/configs"
	ll "github.com/nabla-containers/runnc/llif"
	"github.com/nabla-containers/runnc/runnc-cont"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/vishvananda/netns"
)

var (
	NablaBinDir = "/opt/runnc/bin/"
	NablaRunBin = NablaBinDir + "nabla-run"
)

func newRunncCont(cfg *initConfig) (*runnc_cont.RunncCont, error) {
	if len(cfg.Args) == 0 {
		return nil, fmt.Errorf("OCI process args are empty")
	}

	if !strings.HasSuffix(cfg.Args[0], ".nabla") {
		return nil, fmt.Errorf("entrypoint is not a .nabla file")
	}

	c := runnc_cont.Config{
		NablaRunBin:    NablaRunBin,
		UniKernelBin:   filepath.Join(cfg.Root, cfg.Args[0]),
		Memory:         cfg.Memory,
		Tap:            cfg.TapName,
		IsInKubernetes: true,
		IsInDocker:     false,
		Disk:           []string{cfg.FsPath},
		WorkingDir:     cfg.Cwd,
		Env:            cfg.Env,
		NablaRunArgs:   cfg.Args[1:],
		Mounts:         cfg.Mounts,
	}

	cont, err := runnc_cont.NewRunncCont(c)
	if err != nil {
		return nil, err
	}

	return cont, nil
}

type initConfig struct {
	Id         string          `json:"id"`
	BundlePath string          `json:"bundlepath"`
	Root       string          `json:"root"`
	Args       []string        `json:"args"`
	FsPath     string          `json:"fspath"`
	Cwd        string          `json:"cwd"`
	Env        []string        `json:"env"`
	TapName    string          `json:"tap"`
	NetnsPath  string          `json:"netnspath"`
	Hooks      *spec.Hooks     `json:"hooks"`
	Memory     int64           `json:"mem"`
	Mounts     []spec.Mount    `json:"Mounts"`
	Config     *configs.Config `json:"config"`
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

	// LLC FS Handle
	// TODO(runllc): Add LLStates in here
	fsInput := &ll.FSRunInput{}
	fsInput.ContainerRoot = config.Root
	fsInput.Config = config.Config

	// TODO(runllc): Propagate and store LLstates
	_, err = llcHandler.FSH.FSRunFunc(fsInput)
	if err != nil {
		return fmt.Errorf("Error running llc FS handler: %v", err)
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

	runncCont, err := newRunncCont(config)
	if err != nil {
		return newSystemErrorWithCause(err, "Unable to construct nabla run args")
	}

	if err := runncCont.Run(); err != nil {
		return err
	}

	return nil
}
