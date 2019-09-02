package v1

import (
	"fmt"
	"path"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

type Executor string

const (
	// ExecutorDefault is the name of the ClusterTask that execute kubectl
	ExecutorDefault Executor = "kubectl"
	ExecutorHelm    Executor = "helm"
)

// Config is repo config
type Config struct {
	Manifests []*ManifestOptions `json:"paths,omitempty"`
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

		pipeline.Tasks = append(pipeline.Tasks, tektonv1.PipelineTask{
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
				Name: fmt.Sprintf("alaska-%s-executor", string(executor)),
				Kind: tektonv1.ClusterTaskKind,
			},
		})
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
