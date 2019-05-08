package runnc_cont

import (
	spec "github.com/opencontainers/runtime-spec/specs-go"
)

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
	Gateway   string
	Mac       string

	// Memory max memory size in MBs.
	Memory int64

	// Disk is the path to disk
	Disk []string

	// WorkingDir current working directory.
	WorkingDir string

	// Env is a list of environment variables.
	Env []string

	// Mounts specify source and destination paths that will be copied
	// inside the container's rootfs.
	Mounts []spec.Mount
}
