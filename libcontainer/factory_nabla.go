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
	"os"
	"path/filepath"
	"regexp"
	"syscall"

	"github.com/nabla-containers/runnc/libcontainer/configs"
	ll "github.com/nabla-containers/runnc/llif"

	"github.com/pkg/errors"
)

const (
	stateFilename    = "state.json"
	execFifoFilename = "exec.fifo"
	pauseNablaName   = "pause.nabla"
)

var (
	idRegex  = regexp.MustCompile(`^[\w+-\.]+$`)
	maxIdLen = 1024
)

// New returns a linux based container factory based in the root directory and
// configures the factory with the provided option funcs.
func New(root string, llcHandler ll.RunllcHandler, options ...func(*NablaFactory) error) (Factory, error) {
	if root != "" {
		if err := os.MkdirAll(root, 0700); err != nil {
			return nil, err
		}
	}
	l := &NablaFactory{
		Root:       root,
		LLCHandler: llcHandler,
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
	// LLCHandler is the set of low level container handlers
	LLCHandler ll.RunllcHandler
}

func isPauseContainer(config *configs.Config) bool {
	return len(config.Args) == 1 && config.Args[0] == "/pause"
}

func applyPauseHack(config *configs.Config, containerRoot string) (*configs.Config, error) {
	if !isPauseContainer(config) {
		return nil, errors.New("Trying to make pause changes on non-pause container")
	}

	config.Args = []string{pauseNablaName}
	return config, nil
}

// nablaTapName returns the tapname of a given container ID
func nablaTapName(id string) string {
	if len(id) < 8 {
		panic("Insufficient uniqueness in ID")
	}
	return ("tap" + id)[:syscall.IFNAMSIZ-1]
}

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

	// If it is a pause container for kubernetes, set config so that init
	// will just pause instead of executing a nabla
	fsPath := ""
	if isPauseContainer(config) {
		config, err = applyPauseHack(config, containerRoot)
		if err != nil {
			return nil, err
		}
	} else {
		fshInput := &ll.FSCreateInput{}
		fshInput.ContainerRoot = containerRoot
		fshInput.Config = config
		fshInput.ContainerId = id

		// TODO(runllc): Save llstate in container state
		fsLLState, err := l.LLCHandler.FSH.FSCreateFunc(fshInput)
		if err != nil {
			return nil, err
		}

		// TODO(runllc): Remove this hardcode of fspath when execHandler is in
		fsPath = fsLLState.Options["FsPath"]
	}

	networkhInput := &ll.NetworkCreateInput{}
	networkhInput.ContainerRoot = containerRoot
	networkhInput.Config = config
	networkhInput.ContainerId = id

	// TODO(runllc): Save llstate in container state
	_, err = l.LLCHandler.NetworkH.NetworkCreateFunc(networkhInput)
	if err != nil {
		// TODO(runllc): Handle error case for FS Handler - run FSDestroyFunc
		return nil, fmt.Errorf("Unable to create tap interface: %v", err)
	}

	c := &nablaContainer{
		id:         id,
		root:       containerRoot,
		fsPath:     fsPath,
		config:     config,
		llcHandler: l.LLCHandler,
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
		id:         id,
		root:       containerRoot,
		config:     &state.Config,
		state:      state,
		llcHandler: l.LLCHandler,
	}

	return c, nil
}

func (l *NablaFactory) StartInitialization() error {
	return initNabla(l.LLCHandler)
}

func (l *NablaFactory) Type() string {
	return "nabla"
}

func (l *NablaFactory) validateID(id string) error {
	if !idRegex.MatchString(id) {
		return fmt.Errorf("invalid id format: %v", id)
	}
	if len(id) > maxIdLen {
		return fmt.Errorf("id length: %v, greater than max length: %v", len(id), maxIdLen)
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
