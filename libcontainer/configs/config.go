package configs

type Config struct {
	Args    []string            `json:"args"`
	Rootfs  string              `json:"rootfs"`
	HostUID func() (int, error) `json:"hostuid"`
	HostGID func() (int, error) `json:"hostgid"`
}
