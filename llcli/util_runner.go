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

package llcli

import (
	"github.com/nabla-containers/runnc/libcontainer"
	"github.com/opencontainers/runtime-spec/specs-go"

	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"

	"path/filepath"
)

type runner struct {
	enableSubreaper bool
	shouldDestroy   bool
	detach          bool
	listenFDs       []*os.File
	pidFile         string
	console         string
	container       libcontainer.Container
	create          bool
}

func (r *runner) run(config *specs.Process) (int, error) {
	process, err := newProcess(*config)
	if err != nil {
		r.destroy()
		return -1, err
	}
	if len(r.listenFDs) > 0 {
		process.Env = append(process.Env, fmt.Sprintf("LISTEN_FDS=%d", len(r.listenFDs)), "LISTEN_PID=1")
		process.ExtraFiles = append(process.ExtraFiles, r.listenFDs...)
	}
	rootuid, err := r.container.Config().HostUID()
	if err != nil {
		r.destroy()
		return -1, err
	}
	rootgid, err := r.container.Config().HostGID()
	if err != nil {
		r.destroy()
		return -1, err
	}
	tty, err := setupIO(process, rootuid, rootgid, r.console, config.Terminal, r.detach || r.create)
	if err != nil {
		r.destroy()
		return -1, err
	}
	handler := newSignalHandler(tty, r.enableSubreaper)
	startFn := r.container.Start
	if !r.create {
		startFn = r.container.Run
	}
	defer tty.Close()
	if err := startFn(process); err != nil {
		r.destroy()
		return -1, err
	}
	if err := tty.ClosePostStart(); err != nil {
		r.terminate(process)
		r.destroy()
		return -1, err
	}
	if r.pidFile != "" {
		if err := createPidFile(r.pidFile, process); err != nil {
			r.terminate(process)
			r.destroy()
			return -1, err
		}
	}
	if process.OOMScoreAdj != nil {
		if err := adjustOomScore(process); err != nil {
			r.terminate(process)
			r.destroy()
			return -1, err
		}
	}
	if r.detach || r.create {
		return 0, nil
	}
	status, err := handler.forward(process)
	if err != nil {
		r.terminate(process)
	}
	r.destroy()
	return status, err
}

func (r *runner) destroy() {
	if r.shouldDestroy {
		destroy(r.container)
	}
}

func (r *runner) terminate(p *libcontainer.Process) {
	p.Signal(syscall.SIGKILL)
	p.Wait()
}

// newProcess returns a new libcontainer Process with the arguments from the
// spec and stdio from the current process.
func newProcess(p specs.Process) (*libcontainer.Process, error) {
	lp := &libcontainer.Process{
		Args: p.Args,
		Env:  p.Env,
		// TODO: fix libcontainer's API to better support uid/gid in a typesafe way.
		User:            fmt.Sprintf("%d:%d", p.User.UID, p.User.GID),
		Cwd:             p.Cwd,
		Label:           p.SelinuxLabel,
		NoNewPrivileges: &p.NoNewPrivileges,
		AppArmorProfile: p.ApparmorProfile,
		OOMScoreAdj:     p.OOMScoreAdj,
	}
	for _, gid := range p.User.AdditionalGids {
		lp.AdditionalGroups = append(lp.AdditionalGroups, strconv.FormatUint(uint64(gid), 10))
	}

	return lp, nil
}

// setupIO sets the proper IO on the process depending on the configuration
// If there is a nil error then there must be a non nil tty returned
func setupIO(process *libcontainer.Process, rootuid, rootgid int, console string, createTTY, detach bool) (*tty, error) {
	// detach and createTty will not work unless a console path is passed
	// so error out here before changing any terminal settings
	if createTTY && detach && console == "" {
		return nil, fmt.Errorf("cannot allocate tty if runc will detach")
	}
	if createTTY {
		return createTty(process, rootuid, rootgid, console)
	}
	if detach {
		if err := dupStdio(process, rootuid, rootgid); err != nil {
			return nil, err
		}
		return &tty{}, nil
	}
	return createStdioPipes(process, rootuid, rootgid)
}

// createPidFile creates a file with the processes pid inside it atomically
// it creates a temp file with the paths filename + '.' infront of it
// then renames the file
func createPidFile(path string, process *libcontainer.Process) error {
	pid, err := process.Pid()
	if err != nil {
		return err
	}
	var (
		tmpDir  = filepath.Dir(path)
		tmpName = filepath.Join(tmpDir, fmt.Sprintf(".%s", filepath.Base(path)))
	)
	f, err := os.OpenFile(tmpName, os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%d", pid)
	f.Close()
	if err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func adjustOomScore(process *libcontainer.Process) error {
	pid, err := process.Pid()
	if err != nil {
		return err
	}
	oomScoreAdjPath := filepath.Join("/proc/", strconv.Itoa(pid), "oom_score_adj")
	if process.OOMScoreAdj == nil {
		return fmt.Errorf("OOMScoreAdj value is nil")
	}
	value := strconv.Itoa(*process.OOMScoreAdj)
	err = ioutil.WriteFile(oomScoreAdjPath, []byte(value), 0644)
	if err != nil {
		return err
	}
	return nil
}
