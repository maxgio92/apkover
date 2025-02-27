package pipeline

import (
	"fmt"
	"github.com/maxgio92/apkover/pkg/report"
	"strings"

	"github.com/pkg/errors"

	melange "chainguard.dev/melange/pkg/config"
)

const (
	coverDirGo    = "/tmp/cover"
	envCoverDirGo = "GOCOVERDIR"
)

var (
	errBuildPipelineNotFound = errors.New("build pipeline not found")
)

type GoPipelineUpdater struct{}

func init() {
	registerUpdater("go", new(GoPipelineUpdater))
}

func (u *GoPipelineUpdater) UpdateBuild(build *melange.Configuration) error {
	var found bool
	for _, step := range build.Pipeline {
		// go build as built-in Melange pipeline.
		if step.Uses == "go/build" {
			found = true
			step.With["extra-args"] = "-cover"
		}
		// go build as shell pipeline.
		if strings.Contains(step.Runs, "go build") && !strings.Contains(step.Runs, "-cover") {
			found = true
			strings.Replace(step.Runs, "go build", "go build -cover", 1)
		}
	}
	if !found {
		return errBuildPipelineNotFound
	}

	return nil
}

func (u *GoPipelineUpdater) UpdateTest(test *melange.Test) error {
	// Set the GOCOVERDIR environment variable.
	if test.Environment.Environment == nil {
		test.Environment.Environment = make(map[string]string, 1)
	}
	test.Environment.Environment[envCoverDirGo] = coverDirGo

	// Ensure the coverage data directory exists.
	newPipeline := make([]melange.Pipeline, len(test.Pipeline))
	if !strings.Contains(test.Pipeline[0].Runs, fmt.Sprintf("mkdir -p %s", coverDirGo)) {
		newPipeline = make([]melange.Pipeline, len(test.Pipeline)+1)
		copy(newPipeline[1:], test.Pipeline)

		newPipeline[0] = melange.Pipeline{
			Runs: fmt.Sprintf("mkdir -p %s", coverDirGo),
		}
	} else {
		copy(newPipeline[0:], test.Pipeline)
	}

	// Analyse the coverage with go tool covdata.
	test.Environment.Contents.Packages = append(test.Environment.Contents.Packages, "go")
	command := fmt.Sprintf(`echo "%s $(go tool covdata func -i /tmp/cover  | tail -1 | awk '{print $NF}')"`, report.ReportCovPrefix)
	if !strings.Contains(newPipeline[len(newPipeline)-1].Runs, command) {
		newPipeline = append(newPipeline, melange.Pipeline{
			Runs: command,
		})
	}

	test.Pipeline = newPipeline

	return nil
}
