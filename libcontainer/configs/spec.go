package configs

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

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

	labels := []string{}
	for k, v := range s.Annotations {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}

	netnsPath := ""
	if s.Linux != nil {
		for _, v := range s.Linux.Namespaces {
			if v.Type == specs.NetworkNamespace {
				netnsPath = v.Path
			}
		}
	}

	cfg := Config{
		Args:      s.Process.Args,
		Rootfs:    s.Root.Path,
		Env:       s.Process.Env,
		Cwd:       s.Process.Cwd,
		Version:   s.Version,
		NetnsPath: netnsPath,
		Labels:    labels,
		Hooks:     s.Hooks,
	}

	return &cfg, nil
}
