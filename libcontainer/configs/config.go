package configs

type Config struct {
	Args   []string `json:"args"`
	Rootfs string   `json:"rootfs"`

    // Version is the version of opencontainer specification that is supported.
    Version string `json:"version"`

    // Labels are user defined metadata that is stored in the config and populated on the state
    Labels []string `json:"labels"`
}

// HostUID returns the UID to run the nabla container as. Default is root.
func (c Config) HostUID() (int, error) {
	return 0, nil
}

// HostGID returns the GID to run the nabla container as. Default is root.
func (c Config) HostGID() (int, error) {
	return 0, nil
}
