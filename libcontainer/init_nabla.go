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
)

var (
	NablaBinDir       = "/opt/runnc/bin/"
	NablaRunncContBin = NablaBinDir + "runnc-cont"
	NablaRunBin       = NablaBinDir + "nabla-run"
)

func nablaRunArgs(cfg *initConfig) ([]string, error) {
	if len(cfg.Args) == 0 {
		return nil, fmt.Errorf("OCI process args are empty")
	}

	if !strings.HasSuffix(cfg.Args[0], ".nabla") {
		return nil, fmt.Errorf("Entrypoint is not a .nabla file")
	}

	args := []string{NablaRunncContBin,
		"-nabla-run", NablaRunBin,
		"-tap", cfg.TapName,
		"-cwd", cfg.Cwd,
		"-volume", cfg.FsPath + ":/",
		"-unikernel", filepath.Join(cfg.Root, cfg.Args[0])}

	for _, e := range cfg.Env {
		args = append(args, "-env", e)
	}

	args = append(args, "--")
	args = append(args, cfg.Args[1:]...)

	fmt.Fprintf(os.Stderr, "Running with args: %v", args)
	return args, nil
}

type initConfig struct {
	Root    string   `json:"root"`
	Args    []string `json:"args"`
	FsPath  string   `json:"fspath"`
	Cwd     string   `json:"cwd"`
	Env     []string `json:"env"`
	TapName string   `json:"tap"`
}

func initNabla() error {
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

	runArgs, err := nablaRunArgs(config)
	if err != nil {
		return newSystemErrorWithCause(err, "Unable to construct nabla run args")
	}

	if err := syscall.Exec(runArgs[0], runArgs, os.Environ()); err != nil {
		return err
	}

	return nil

}
