package runnc_cont

// Config configuration to create a runnc-cont
type Config struct {
	// NablaRunBin is the path to 'nabla-run' binary.
	NablaRunBin string

	NablaRunArgs []string

	// UniKernelBin is the path to 'unikernel' binary.
	UniKernelBin string

	// Tap tap device. (e.g. tap100)
	Tap string

	IPAddress string
	IPMask    int
	GateWay   string

	// Memory max memory size in MBs.
	Memory int64

	// IsInDocker means running in a Docker container or not.
	IsInDocker bool

	// IsInKubernetes means running in a Kubernetes cluster or not.
	IsInKubernetes bool

	// Disk is the path to disk
	Disk []string

	// WorkingDir current working directory.
	WorkingDir string

	// Env is a list of environment variables.
	Env []string
}
