// +build linux

package libcontainer

import (
	"fmt"
	"github.com/nabla-containers/runnc/libcontainer/configs"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
)

const (
	stateFilename    = "state.json"
	execFifoFilename = "exec.fifo"
)

var (
	idRegex  = regexp.MustCompile(`^[\w+-\.]+$`)
	maxIdLen = 1024
)

// New returns a linux based container factory based in the root directory and
// configures the factory with the provided option funcs.
// TODO(NABLA)
func New(root string, options ...func(*NablaFactory) error) (Factory, error) {
	if root != "" {
		if err := os.MkdirAll(root, 0700); err != nil {
			return nil, err
		}
	}
	l := &NablaFactory{
		Root: root,
		//InitArgs: []string{"/proc/self/exe", "init"},
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
	//InitArgs []string
}

// TODO(NABLA)
func (l *NablaFactory) Create(id string, config *configs.Config) (Container, error) {
	if l.Root == "" {
		return nil, fmt.Errorf("invalid root")
	}
	if err := l.validateID(id); err != nil {
		return nil, err
	}
	//if err := l.Validator.Validate(config); err != nil {
	//    return nil, err
	//}
	uid, err := config.HostUID()
	if err != nil {
		return nil, err
	}
	gid, err := config.HostGID()
	if err != nil {
		return nil, err
	}
	containerRoot := filepath.Join(l.Root, id)
	if _, err := os.Stat(containerRoot); err == nil {
		return nil, fmt.Errorf("container with id exists: %v", id)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if err := os.MkdirAll(containerRoot, 0711); err != nil {
		return nil, err
	}

	if err := os.Chown(containerRoot, uid, gid); err != nil {
		return nil, err
	}
	fifoName := filepath.Join(containerRoot, execFifoFilename)
	oldMask := syscall.Umask(0000)
	if err := syscall.Mkfifo(fifoName, 0622); err != nil {
		syscall.Umask(oldMask)
		return nil, err
	}
	syscall.Umask(oldMask)
	if err := os.Chown(fifoName, uid, gid); err != nil {
		return nil, err
	}

	c := &nablaContainer{
		id:     id,
		root:   containerRoot,
		config: config,
		state:  Stopped,
	}
	return c, nil
}

// TODO(NABLA)
func (l *NablaFactory) Load(id string) (Container, error) {
	return nil, errors.New("NablaFactory.Load not implemented")
}

// TODO(NABLA)
func (l *NablaFactory) StartInitialization() error {
	return errors.New("NablaFactory.StartInitialization not implemented")
}

func (l *NablaFactory) Type() string {
	return "nabla"
}

func (l *NablaFactory) validateID(id string) error {
	if !idRegex.MatchString(id) {
		return fmt.Errorf("invalid id format: %v", id)
	}
	if len(id) > maxIdLen {
		return fmt.Errorf("invalid id format: %v", id)
	}
	return nil
}
