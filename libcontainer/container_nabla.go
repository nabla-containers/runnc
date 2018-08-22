// +build linux

package libcontainer

import (
	"fmt"
	"github.com/nabla-containers/runnc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/system"
	"github.com/opencontainers/runc/libcontainer/utils"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const stdioFdCount = 3

// State represents a running container's state
type State struct {
	BaseState

	// Platform specific fields below here
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
	id     string
	root   string
	config *configs.Config
	//cgroupManager        cgroups.Manager
	//initArgs             []string
	//initProcess          parentProcess
	//initProcessStartTime string
	//criuPath             string
	m sync.Mutex
	//criuVersion          int
	status  Status
	state   State
	created time.Time
}

func (c *nablaContainer) Config() configs.Config {
	return *c.config
}

// TODO(NABLA)
func (c *nablaContainer) Status() (Status, error) {
	c.m.Lock()
	defer c.m.Unlock()
	var sts Status
	return sts, errors.New("NablaContainer.Status not implemented")
}

// TODO(NABLA)
func (c *nablaContainer) State() (*State, error) {
	c.m.Lock()
	defer c.m.Unlock()
	return nil, errors.New("NablaContainer.State not implemented")
}

// TODO(NABLA)
func (c *nablaContainer) Destroy() error {
	c.m.Lock()
	defer c.m.Unlock()
	return errors.New("NablaContainer.Destroy not implemented")
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

// TODO(NABLA)
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
	return errors.New("NablaContainer.Exec not implemented")
}

// TODO(NABLA)
func (c *nablaContainer) Signal(s os.Signal) error {
	return errors.New("NablaContainer.Signal not implemented")
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

func (c *nablaContainer) start(p *Process) error {
	cmd := exec.Command(p.Args[0], p.Args[1:]...)
	cmd.Stdin = p.Stdin
	cmd.Stdout = p.Stdout
	cmd.Stderr = p.Stderr
	cmd.Dir = c.config.Rootfs
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.ExtraFiles = p.ExtraFiles
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("_LIBCONTAINER_INITPIPE=%d", stdioFdCount+len(cmd.ExtraFiles)-2),
		fmt.Sprintf("_LIBCONTAINER_STATEDIR=%d", stdioFdCount+len(cmd.ExtraFiles)-1))

	if err := cmd.Start(); err != nil {
		return err
	}

	if cmd.Process == nil {
		return errors.New("Cmd.Process is nil after starting")
	}

	p.ops = &nablaProcess{
		process: p,
		cmd:     cmd,
	}

	// TODO: Create state  and update state JSON
	var err error
	c.status = Created
	c.state.BaseState.InitProcessPid = p.ops.pid()
	c.state.BaseState.Created = time.Now().UTC()
	c.state.BaseState.InitProcessStartTime, err = system.GetProcessStartTime(c.state.BaseState.InitProcessPid)
	if err != nil {
		return err
	}

	c.saveState(&c.state)

	return nil
}

func (c *nablaContainer) saveState(s *State) error {
	f, err := os.Create(filepath.Join(c.root, stateFilename))
	if err != nil {
		return err
	}
	defer f.Close()
	return utils.WriteJSON(f, s)
}
