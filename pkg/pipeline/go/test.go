package gopipeline

import (
	melange "chainguard.dev/melange/pkg/config"
	"fmt"
	"strings"
)

func UpdateTest(test *melange.Test) error {
	// Set the GOVERDIR environment variable.
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
	command := `echo "apkoverage: $(go tool covdata func -i /tmp/cover  | tail -1 | awk '{print $NF}')"`
	if !strings.Contains(newPipeline[len(newPipeline)-1].Runs, command) {
		newPipeline = append(newPipeline, melange.Pipeline{
			Runs: command,
		})
	}

	test.Pipeline = newPipeline

	return nil
}
