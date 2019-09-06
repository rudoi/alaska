package v1

import (
	"fmt"
	"path"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

type Executor string

const (
	// ExecutorDefault is the executor that executes kubectl
	ExecutorDefault Executor = "kubectl"

	// ExecutorHelm is the executor that executes helm
	ExecutorHelm Executor = "helm"

	// Executor Task name format string
	ExecutorTaskNameFormatString = "alaska-%s-executor"
)

type Strategy string

const (
	StrategyDefault    Strategy = "parallel"
	StrategySequential Strategy = "sequential"
)

// Config is repo config
type Config struct {
	Manifests []*ManifestOptions `json:"paths,omitempty"`
	Strategy  Strategy           `json:"strategy,omitempty"`
}

// ManifestOptions describes the path to a manifest and its type
type ManifestOptions struct {
	Path string   `json:"path,omitempty"`
	Type Executor `json:"type,omitempty"`
}

func (c *Config) ToPipelineSpec() tektonv1.PipelineSpec {
	pipeline := tektonv1.PipelineSpec{
		Resources: []tektonv1.PipelineDeclaredResource{
			{
				Name: "repo",
				Type: tektonv1.PipelineResourceTypeGit,
			},
			{
				Name: "cluster",
				Type: tektonv1.PipelineResourceTypeCluster,
			},
		},
		Tasks: []tektonv1.PipelineTask{},
	}

	for i, manifest := range c.Manifests {
		var executor Executor
		if manifest.Type == "" {
			executor = ExecutorDefault
		} else {
			executor = Executor(manifest.Type)
		}

		task := tektonv1.PipelineTask{
			Name:   fmt.Sprintf("task-%d", i),
			Params: manifest.ToParams(),
			Resources: &tektonv1.PipelineTaskResources{
				Inputs: []tektonv1.PipelineTaskInputResource{
					{
						Name:     "repo",
						Resource: "repo",
					},
					{
						Name:     "cluster",
						Resource: "cluster",
					},
				},
			},
			TaskRef: tektonv1.TaskRef{
				Name: executor.toTaskName(),
				Kind: tektonv1.ClusterTaskKind,
			},
		}

		if c.Strategy == StrategySequential && i > 0 {
			task.RunAfter = []string{fmt.Sprintf("task-%d", i-1)}
		}

		pipeline.Tasks = append(pipeline.Tasks, task)
	}
	return pipeline
}

func (mo *ManifestOptions) ToParams() (params []tektonv1.Param) {
	params = append(params, tektonv1.Param{
		Name: "path",
		Value: tektonv1.ArrayOrString{
			Type:      tektonv1.ParamTypeString,
			StringVal: mo.Path,
		},
	})

	if mo.Type == ExecutorHelm {
		params = append(params, tektonv1.Param{
			Name: "release",
			Value: tektonv1.ArrayOrString{
				Type:      tektonv1.ParamTypeString,
				StringVal: path.Base(mo.Path),
			},
		})
	}

	return
}

func (e Executor) toTaskName() string {
	return fmt.Sprintf(ExecutorTaskNameFormatString, string(e))
}
