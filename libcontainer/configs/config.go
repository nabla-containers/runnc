package configs

type Config struct {
	Args   []string `json:"args"`
	Rootfs string   `json:"rootfs"`
}

// HostUID returns the UID to run the nabla container as. Default is root.
func (c Config) HostUID() (int, error) {
	return 0, nil
}

// HostGID returns the GID to run the nabla container as. Default is root.
func (c Config) HostGID() (int, error) {
	return 0, nil
}
