// +build linux

package libcontainer

import (
	"encoding/json"
	"fmt"
	"github.com/nabla-containers/runnc/libcontainer/configs"
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
		state: &State{
			BaseState: BaseState{
				ID:     id,
				Config: *config,
			},
			Status: Stopped,
		},
	}
	return c, nil
}

// TODO(NABLA)
func (l *NablaFactory) Load(id string) (Container, error) {
	if l.Root == "" {
		return nil, newGenericError(fmt.Errorf("invalid root"), ConfigInvalid)
	}
	containerRoot := filepath.Join(l.Root, id)
	state, err := l.loadState(containerRoot, id)
	if err != nil {
		return nil, err
	}

	c := &nablaContainer{
		id:     id,
		root:   containerRoot,
		config: &state.Config,
		state:  state,
	}

	return c, nil
}

// TODO(NABLA)
func (l *NablaFactory) StartInitialization() error {
	return initNabla()
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

func (l *NablaFactory) loadState(root, id string) (*State, error) {
	f, err := os.Open(filepath.Join(root, stateFilename))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, newGenericError(fmt.Errorf("container %q does not exist", id), ContainerNotExists)
		}
		return nil, newGenericError(err, SystemError)
	}
	defer f.Close()
	var state *State
	if err := json.NewDecoder(f).Decode(&state); err != nil {
		return nil, newGenericError(err, SystemError)
	}
	return state, nil
}
