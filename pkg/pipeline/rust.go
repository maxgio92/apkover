package pipeline

import (
	"fmt"
	"strings"

	melange "chainguard.dev/melange/pkg/config"

	"github.com/maxgio92/apkover/pkg/report"
)

const (
	envRustFlags = "RUSTFLAGS"
	rustFlags    = "-C instrument-coverage"
)

var (
	testDeps   = []string{"clang", "rust"}
	testCovCmd = fmt.Sprintf(`llvm-profdata merge \
  --sparse default_*.profraw \
  --output default.profdata 2>/dev/null
echo -n "%s "
llvm-cov report \
  --ignore-filename-regex='/.cargo/registry' \
  --instr-profile=default.profdata \
  --object "/$(apk info -L ${{package.name}} 2>/dev/null | grep 'bin\/' | head -1)" \
  --summary-only 2>/dev/null \
  | tail -1 | awk '{print $7}'`, report.ReportCovPrefix)
)

type RustPipelineUpdater struct{}

func init() {
	registerUpdater("rust", new(RustPipelineUpdater))
}

func (u *RustPipelineUpdater) UpdateBuild(build *melange.Configuration) error {
	if build.Environment.Environment == nil {
		build.Environment.Environment = make(map[string]string, 1)
	}
	build.Environment.Environment[envRustFlags] = rustFlags

	return nil
}

func (u *RustPipelineUpdater) UpdateTest(test *melange.Test) error {
	// Update test time dependencies.
	if test.Environment.Contents.Packages == nil {
		test.Environment.Contents.Packages = make([]string, 2)
	}
	test.Environment.Contents.Packages = append(test.Environment.Contents.Packages, testDeps...)

	// Merge coverage data with llvm-profdata merge and report coverage with llvm-cov report.
	// Docs: https://doc.rust-lang.org/rustc/instrument-coverage.html#running-the-instrumented-binary-to-generate-raw-coverage-profiling-data.
	test.Environment.Contents.Packages = append(test.Environment.Contents.Packages, "go")
	if !strings.Contains(test.Pipeline[len(test.Pipeline)-1].Runs, testCovCmd) {
		test.Pipeline = append(test.Pipeline, melange.Pipeline{
			Runs: testCovCmd,
		})
	}

	return nil
}
