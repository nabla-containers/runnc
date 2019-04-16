package llif

import (
	"github.com/nabla-containers/runnc/libcontainer/configs"
)

type FsGenericInput struct {
	// ContainerId is the id of the container
	ContainerId string

	// ContainerRoot signifies the root of the container's existence on the
	// host
	ContainerRoot string

	// Config contains the configuration of the container
	Config *configs.Config

	// The state of LL handlers
	FsState      *LLState
	NetworkState *LLState
	ExecState    *LLState
}

type FsCreateInput struct {
	FsGenericInput
}

type FsRunInput struct {
	FsGenericInput
}

type FsDestroyInput struct {
	FsGenericInput
}
