package configs

import (
	spec "github.com/opencontainers/runtime-spec/specs-go"
)

// Config represents the configuration necessary for nabla containers from the OCI runtime spec
type Config struct {
	Args   []string `json:"args"`
	Rootfs string   `json:"rootfs"`
	Env    []string `json:"env"`
	Cwd    string   `json:"cwd"`

	// Version is the version of opencontainer specification that is supported.
	Version string `json:"version"`

	// Labels are user defined metadata that is stored in the config and populated on the state
	Labels []string `json:"labels"`

	// Network namespace
	NetnsPath string `json:"netnspath"`

	// Hooks configures callbacks for container lifecycle events.
	Hooks *spec.Hooks `json:"hooks,omitempty"`
}

// HostUID returns the UID to run the nabla container as. Default is root.
func (c Config) HostUID() (int, error) {
	return 0, nil
}

// HostGID returns the GID to run the nabla container as. Default is root.
func (c Config) HostGID() (int, error) {
	return 0, nil
}
