package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/testscript"
)

const (
	dirE2e      = "e2e"
	dirTestData = "testdata"
	name        = "apkover"
)

func TestApkover(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:                 dirE2e,
		RequireExplicitExec: true,
		Setup: func(env *testscript.Env) error {
			// Prepare testdata.
			err := copyDir(filepath.Join(dirE2e, dirTestData), filepath.Join(env.WorkDir, dirTestData))
			if err != nil {
				return errors.Wrap(err, "error preparing testdata")
			}

			homeDir := filepath.Join(env.WorkDir, "home")
			if err := os.MkdirAll(homeDir, 0755); err != nil {
				return errors.Wrap(err, "error creating the test home dir")
			}
			env.Setenv("HOME", homeDir)

			// Build the binary instrumented for integration tests coverage.
			path := filepath.Join(env.WorkDir, name)
			cmd := exec.Command("go", "build", "-v", "-o", path, ".")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				return errors.Wrap(err, "error building the binary for tests")
			}
			// Add the binary path to PATH.
			fmt.Println(env.WorkDir)
			newPath := env.WorkDir + string(os.PathListSeparator) + os.Getenv("PATH")
			env.Setenv("PATH", newPath)

			// Set the coverage data directory.
			env.Setenv("GOCOVERDIR", "/tmp/cover")

			return nil
		},
	})
}

func copyDir(src string, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
			continue
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}

		err = os.WriteFile(dstPath, data, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
