// +build linux

package libcontainer

import (
	"github.com/nabla-containers/runnc/libcontainer/configs"
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
