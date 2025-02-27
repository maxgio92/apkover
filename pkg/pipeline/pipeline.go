package pipeline

import (
	melange "chainguard.dev/melange/pkg/config"
)

var pipelineUpdates = make(map[string]PipelineUpdater)

func registerUpdater(name string, update PipelineUpdater) {
	pipelineUpdates[name] = update
}

func GetUpdater(name string) PipelineUpdater {
	return pipelineUpdates[name]
}

type PipelineUpdater interface {
	UpdateBuild(build *melange.Configuration) error
	UpdateTest(test *melange.Test) error
}
