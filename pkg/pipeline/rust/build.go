package rust

import (
	melange "chainguard.dev/melange/pkg/config"
)

const (
	envRustFlags = "RUSTFLAGS"
	rustFlags    = "-C instrument-coverage"
)

func UpdateBuild(build *melange.Configuration) error {
	if build.Environment.Environment == nil {
		build.Environment.Environment = make(map[string]string, 1)
	}
	build.Environment.Environment[envRustFlags] = rustFlags

	return nil
}
