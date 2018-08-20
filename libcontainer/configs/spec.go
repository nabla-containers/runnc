package configs

import (
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

// TODO(NABLA)
func ParseSpec(s *specs.Spec) (*Config, error) {
	if s == nil {
		return nil, errors.New("Spec is nil")
	}
	if s.Process == nil || s.Process.Args == nil {
		return nil, errors.New("Process Args is nil")
	}

	if s.Root == nil {
		return nil, errors.New("Root is nil")
	}

	cfg := Config{
		Args:   s.Process.Args,
		Rootfs: s.Root.Path,
	}

	return &cfg, nil
}
