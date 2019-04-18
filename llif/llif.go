package llif

// RunllcHandler is the interface that is needed to be implemented in order
// to create a Low Level OCI runtime with Runllc.
//
// There are 3 extensible components and 3 integration points.
//
// The 3 extensible components are the filesystem, network, and execution.
// Thus, there are 3 separate handles for each of them. Fs, Network, Exec.
//
// There are 3 different integration points, creation of the container,
// Running of the container, and finally, the destruction of the container.
//
// The order of which the handlers are run are as follows:
// Integration: Create
// Order: FsCreateFunc, NetworkCreateFunc, ExecCreateFunc
//
// Integration: Run
// Order: FsRunFunc, NetworkRunFunc, ExecRunFunc
//
// Integration: Destroy (this is the backward order from the previous two)
// Order: ExecDestroyFunc, NetworkDestroyFunc, FsDestroyFunc
type RunllcHandler struct {
	FsH      FsHandler
	NetworkH NetworkHandler
	ExecH    ExecHandler
}

type FsHandler interface {
	FsCreateFunc(*FsCreateInput) (*LLState, error)
	FsRunFunc(*FsRunInput) (*LLState, error)
	FsDestroyFunc(*FsDestroyInput) (*LLState, error)
}

type NetworkHandler interface {
	NetworkCreateFunc(*NetworkCreateInput) (*LLState, error)
	NetworkRunFunc(*NetworkRunInput) (*LLState, error)
	NetworkDestroyFunc(*NetworkDestroyInput) (*LLState, error)
}

type ExecHandler interface {
	ExecCreateFunc(*ExecCreateInput) (*LLState, error)
	// ExecRunFunc should not return unless it runs into an error
	// TODO(runllc): Change this to possibly return state
	ExecRunFunc(*ExecRunInput) error
	ExecDestroyFunc(*ExecDestroyInput) (*LLState, error)
}

type LLState struct {
	// Options is the map of parameters that will be stored in the config and
	// passed along across different operations. Entries in this map set
	// in the output of the Create phase will be present in the input of the
	// Run phase.
	// TODO(runllc): Need to figure out how to save state in Exec phase or is
	// there a need to?
	Options map[string]string `json:"options"`

	// InMemoryObjects is the map of objects that can be shared with other handlers
	// within the same operation (i.e. in-memory data structures). The entries
	// from the output of the Create phase will not be accessible to the
	// Run phase. However, they will be accessible by the Exec handler of the
	// same phase.
	InMemoryObjects map[string]interface{} `json:"-"`
}
