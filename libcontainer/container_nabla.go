// +build linux

package libcontainer

import (
	"github.com/nabla-containers/runnc/libcontainer/configs"
	"github.com/pkg/errors"
	"os"
	"sync"
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
	state   Status
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
	return errors.New("NablaContainer.Start not implemented")
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
