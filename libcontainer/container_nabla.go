// +build linux

package libcontainer

import (
	"encoding/json"
	"fmt"
	"github.com/nabla-containers/runnc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/system"
	"github.com/opencontainers/runc/libcontainer/utils"
	"github.com/pkg/errors"
	"io/ioutil"
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
	state   *State
	created time.Time
}

func (c *nablaContainer) Config() configs.Config {
	return *c.config
}

// TODO(NABLA)
func (c *nablaContainer) Status() (Status, error) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.currentStatus()
}

// TODO(NABLA)
func (c *nablaContainer) State() (*State, error) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.currentState()
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

	// TODO: Write config to pipe for child to receive JSON
	defer parentPipe.Close()
	config := initConfig{
		Args: c.config.Args,
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

	// TODO: Create state  and update state JSON
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
	// TODO: refreshState (by looking at system info and verifying state)
	return c.state, nil
}

func (c *nablaContainer) currentStatus() (Status, error) {
	return c.state.Status, nil
}

func (c *nablaContainer) saveState(s *State) error {
	f, err := os.Create(filepath.Join(c.root, stateFilename))
	if err != nil {
		return err
	}
	defer f.Close()
	return utils.WriteJSON(f, s)
}
