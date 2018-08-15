// +build linux

package libcontainer

import (
	"github.com/nabla-containers/runnc/libcontainer/configs"
	"os"
	"regexp"
)

const (
	stateFilename    = "state.json"
	execFifoFilename = "exec.fifo"
)

var (
	idRegex  = regexp.MustCompile(`^[\w+-\.]+$`)
	maxIdLen = 1024
)

// TODO(NABLA)
// New returns a linux based container factory based in the root directory and
// configures the factory with the provided option funcs.
func New(root string, options ...func(*NablaFactory) error) (Factory, error) {
	if root != "" {
		if err := os.MkdirAll(root, 0700); err != nil {
			return nil, err
		}
	}
	l := &NablaFactory{
		Root:     root,
		InitArgs: []string{"/proc/self/exe", "init"},
	}

	for _, opt := range options {
		if err := opt(l); err != nil {
			return nil, err
		}
	}
	return l, nil
}

// LinuxFactory implements the default factory interface for linux based systems.
type NablaFactory struct {
	// Root directory for the factory to store state.
	Root string

	// InitArgs are arguments for calling the init responsibilities for spawning
	// a container.
	InitArgs []string
}

// TODO(NABLA)
func (l *NablaFactory) Create(id string, config *configs.Config) (Container, error) {
	return nil, nil
}

// TODO(NABLA)
func (l *NablaFactory) Load(id string) (Container, error) {
	return nil, nil
}

// TODO(NABLA)
func (l *NablaFactory) StartInitialization() error {
	return nil
}

func (l *NablaFactory) Type() string {
	return "nabla"
}
