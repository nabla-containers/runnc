package llif

// RuncllcHandler is the interface that is needed to be implemented in order
// to create a Low Level OCI runtime with Runllc.
//
// There are 3 extensible components and 3 integration points.
//
// The 3 extensible components are the filesystem, network, and execution.
// Thus, there are 3 separate handles for each of them. FS, Network, Exec.
//
// There are 3 different integration points, creation of the container,
// Running of the container, and finally, the destruction of the container.
//
// The order of which the handlers are run are as follows:
// Integration: Create
// Order: FSCreateFunc, NetworkCreateFunc, ExecCreateFunc
//
// Integration: Run
// Order: FSRunFunc, NetworkRunFunc, ExecRunFunc
//
// Integration: Destroy (this is the backward order from the previous two)
// Order: ExecDestroyFunc, NetworkDestroyFunc, FSDestroyFunc
type RunllcHandler interface {
	FSHandler
	NetworkHandler
	ExecHandler
}

type FSHandler interface {
	FSCreateFunc(*FSCreateInput) (FSCreateOutput, error)
	FSRunFunc(*FSRunInput) (FSRunOutput, error)
	FSDestroyFunc(*FSDestroyInput) (FSDestroyOutput, error)
}

type NetworkHandler interface {
	NetworkCreateFunc(*NetworkCreateInput) (NetworkCreateOutput, error)
	NetworkRunFunc(*NetworkRunInput) (NetworkRunOutput, error)
	NetworkDestroyFunc(*NetworkDestroyInput) (NetworkDestroyOutput, error)
}

type ExecHandler interface {
	ExecCreateFunc(*ExecCreateInput) (ExecCreateOutput, error)
	ExecRunFunc(*ExecRunInput) (ExecRunOutput, error)
	ExecDestroyFunc(*ExecDestroyInput) (ExecDestroyOutput, error)
}
