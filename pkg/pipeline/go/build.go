package gopipeline

import (
	"strings"

	melange "chainguard.dev/melange/pkg/config"
)

func UpdateBuild(pipeline []melange.Pipeline) error {
	var found bool
	for _, step := range pipeline {
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
