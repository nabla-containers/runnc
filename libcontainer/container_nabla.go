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

// +build linux

package libcontainer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/nabla-containers/runnc/libcontainer/configs"
	ll "github.com/nabla-containers/runnc/llif"
	"github.com/opencontainers/runc/libcontainer/system"
	"github.com/opencontainers/runc/libcontainer/utils"
	"github.com/pkg/errors"
)

const stdioFdCount = 3

// State represents a running container's state
type State struct {
	BaseState

	FsState      ll.LLState `json:"fsstate"`
	NetworkState ll.LLState `json:"netstate"`
	ExecState    ll.LLState `json:"execstate"`

	// Platform specific fields below here
	Status Status `json:"status"`
}

// Container is a libcontainer container object.
//
// Each container is thread-safe within the same process. Since a container can
// be destroyed by a separate process, any function may return that the container
// was not found.
type Container interface {
	BaseContainer

	// Methods below here are platform specific
}

type nablaContainer struct {
	id         string
	root       string
	config     *configs.Config
	m          sync.Mutex
	state      *State
	created    time.Time
	llcHandler ll.RunllcHandler
}

func (c *nablaContainer) Config() configs.Config {
	return *c.config
}

func (c *nablaContainer) Status() (Status, error) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.currentStatus()
}

func (c *nablaContainer) State() (*State, error) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.currentState()
}

func (c *nablaContainer) Destroy() error {
	c.m.Lock()
	defer c.m.Unlock()
	return c.destroy()
}

func (c *nablaContainer) ID() string {
	return c.id
}

// TODO(NABLA)
func (c *nablaContainer) Processes() ([]int, error) {
	return nil, errors.New("NablaContainer.Processes not implemented")
}

// TODO(NABLA)
func (c *nablaContainer) Stats() (*Stats, error) {
	return nil, errors.New("NablaContainer.Stats not implemented")
}

// TODO(NABLA)
func (c *nablaContainer) Set(config configs.Config) error {
	c.m.Lock()
	defer c.m.Unlock()
	return errors.New("NablaContainer.Set not implemented")
}

func (c *nablaContainer) Start(process *Process) error {
	c.m.Lock()
	defer c.m.Unlock()
	return c.start(process)
}

// TODO(NABLA)
func (c *nablaContainer) Run(process *Process) error {
	c.m.Lock()
	defer c.m.Unlock()
	return errors.New("NablaContainer.Run not implemented")
}

// TODO(NABLA)
func (c *nablaContainer) Exec() error {
	c.m.Lock()
	defer c.m.Unlock()
	return c.exec()
}

func (c *nablaContainer) Signal(sig os.Signal, all bool) error {
	// For nabla container, we only have 1 process
	s, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New("os: unsupported signal type")
	}
	pid := c.state.InitProcessPid
	if all {
		pid = -pid
	}
	return syscall.Kill(pid, s)
}

type nablaProcess struct {
	process *Process
	cmd     *exec.Cmd
}

func (p *nablaProcess) wait() (*os.ProcessState, error) {
	err := p.cmd.Wait()
	return p.cmd.ProcessState, err
}

func (p *nablaProcess) pid() int {
	return p.cmd.Process.Pid
}

func (p *nablaProcess) signal(sig os.Signal) error {
	s, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New("os: unsupported signal type")
	}
	return syscall.Kill(p.pid(), s)
}

// TODO: TEMP COPY OUTDATED VERSION OF RUNC
// NewSockPair returns a new unix socket pair
func NewSockPair(name string) (parent *os.File, child *os.File, err error) {
	fds, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, nil, err
	}
	return os.NewFile(uintptr(fds[1]), name+"-p"), os.NewFile(uintptr(fds[0]), name+"-c"), nil
}

func (c *nablaContainer) start(p *Process) error {
	parentPipe, childPipe, err := NewSockPair("init")
	if err != nil {
		return newSystemErrorWithCause(err, "creating new init pipe")
	}
	cmd, err := c.commandTemplate(p, childPipe)
	if err != nil {
		return newSystemErrorWithCause(err, "creating new command template")
	}

	// We only set up rootDir if we're not doing a `runc exec`. The reason for
	// this is to avoid cases where a racing, unprivileged process inside the
	// container can get access to the statedir file descriptor (which would
	// allow for container rootfs escape).
	rootDir, err := os.Open(c.root)
	if err != nil {
		return err
	}
	cmd.ExtraFiles = append(cmd.ExtraFiles, rootDir)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("_LIBCONTAINER_STATEDIR=%d", stdioFdCount+len(cmd.ExtraFiles)-1))

	// newInitProcess
	p.ops = &nablaProcess{
		process: p,
		cmd:     cmd,
	}

	defer parentPipe.Close()
	config := initConfig{
		Id:           c.id,
		BundlePath:   c.root,
		Root:         c.config.Rootfs,
		Args:         c.config.Args,
		Cwd:          c.config.Cwd,
		Env:          c.config.Env,
		NetnsPath:    c.config.NetnsPath,
		Hooks:        c.config.Hooks,
		Memory:       c.config.Memory,
		Mounts:       c.config.Mounts,
		Config:       c.config,
		FsState:      c.state.FsState,
		NetworkState: c.state.NetworkState,
		ExecState:    c.state.ExecState,
	}

	enc := json.NewEncoder(parentPipe)
	if err := enc.Encode(config); err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if cmd.Process == nil {
		return errors.New("Cmd.Process is nil after starting")
	}

	c.state.InitProcessPid = p.ops.pid()
	c.state.Created = time.Now().UTC()
	c.state.Status = Created
	c.state.InitProcessStartTime, err = system.GetProcessStartTime(c.state.BaseState.InitProcessPid)
	if err != nil {
		return err
	}

	c.saveState(c.state)

	return nil
}

func (c *nablaContainer) exec() error {
	path := filepath.Join(c.root, execFifoFilename)
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return newSystemErrorWithCause(err, "open exec fifo for reading")
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if len(data) > 0 {
		c.state.Status = Running
		c.saveState(c.state)
		os.Remove(path)
		return nil
	}
	return fmt.Errorf("cannot start an already running container")
}

func (c *nablaContainer) destroy() error {
	c.state.InitProcessPid = 0
	c.state.Status = Stopped

	fsInput := &ll.FSDestroyInput{}
	fsInput.ContainerRoot = c.root
	fsInput.Config = c.config
	fsInput.ContainerId = c.id
	fsInput.FSState = &c.state.FsState
	fsInput.NetworkState = &c.state.NetworkState
	fsInput.ExecState = &c.state.ExecState

	fsState, err := c.llcHandler.FSH.FSDestroyFunc(fsInput)
	if err != nil {
		return err
	}
	if fsState != nil {
		c.state.FsState = *fsState
	} else {
		c.state.FsState = ll.LLState{}
	}

	networkInput := &ll.NetworkDestroyInput{}
	networkInput.ContainerRoot = c.root
	networkInput.Config = c.config
	networkInput.ContainerId = c.id
	networkInput.FSState = fsState
	networkInput.NetworkState = &c.state.NetworkState
	networkInput.ExecState = &c.state.ExecState

	networkState, err := c.llcHandler.NetworkH.NetworkDestroyFunc(networkInput)
	if err != nil {
		return err
	}

	if networkState != nil {
		c.state.NetworkState = *networkState
	} else {
		c.state.NetworkState = ll.LLState{}
	}

	return nil
}

func (c *nablaContainer) commandTemplate(p *Process, childPipe *os.File) (*exec.Cmd, error) {
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.Stdin = p.Stdin
	cmd.Stdout = p.Stdout
	cmd.Stderr = p.Stderr
	cmd.Dir = c.config.Rootfs
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmd.ExtraFiles = append(cmd.ExtraFiles, p.ExtraFiles...)
	cmd.ExtraFiles = append(cmd.ExtraFiles, childPipe)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("_LIBCONTAINER_INITPIPE=%d", stdioFdCount+len(cmd.ExtraFiles)-1),
	)
	// NOTE: when running a container with no PID namespace and the parent process spawning the      container is
	// PID1 the pdeathsig is being delivered to the container's init process by the kernel for some  reason
	// even with the parent still running.
	/* TODO: Check if needed
	   if c.config.ParentDeathSignal > 0 {
	       cmd.SysProcAttr.Pdeathsig = syscall.Signal(c.config.ParentDeathSignal)
	   }
	*/
	return cmd, nil
}

func (c *nablaContainer) currentState() (*State, error) {
	process, err := os.FindProcess(int(c.state.InitProcessPid))
	if err != nil {
		return nil, err
	}

	// [kill(2)]  If  sig  is 0, then no signal is sent, but error checking is still per‐
	// formed; this can be used to check for the existence of a process ID  or
	// process group ID.
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		c.state.Status = Stopped
		c.saveState(c.state)
	}

	return c.state, nil
}

func (c *nablaContainer) currentStatus() (Status, error) {
	var sts Status
	state, err := c.currentState()
	if err != nil {
		return sts, err
	}

	return state.Status, nil
}

func (c *nablaContainer) saveState(s *State) error {
	f, err := os.Create(filepath.Join(c.root, stateFilename))
	if err != nil {
		return err
	}
	defer f.Close()
	return utils.WriteJSON(f, s)
}
