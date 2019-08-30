package alaska

import (
	"fmt"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

type Executor string

const (
	// ExecutorDefault is the name of the ClusterTask that execute kubectl
	ExecutorDefault Executor = "alaska-kubectl-executor"
)

// Config is some config
type Config struct {
	Paths []string `json:"paths,omitempty"`
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

	for i, path := range c.Paths {
		pipeline.Tasks = append(pipeline.Tasks, tektonv1.PipelineTask{
			Name: fmt.Sprintf("task-%d", i),
			Params: []tektonv1.Param{
				{
					Name: "path",
					Value: tektonv1.ArrayOrString{
						Type:      tektonv1.ParamTypeString,
						StringVal: path,
					},
				},
			},
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
				Name: string(ExecutorDefault),
				Kind: tektonv1.ClusterTaskKind,
			},
		})
	}
	return pipeline
}
