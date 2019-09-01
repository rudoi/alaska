package alaska

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

var _ = Describe("Config.ToPipelineSpec tests", func() {
	Context("given a config with a valid path", func() {
		var cfg *Config

		BeforeEach(func() {
			cfg = &Config{
				Paths: []string{"test.yaml"},
			}
		})

		It("should return a pipeline spec with one task", func() {
			expected := tektonv1.PipelineSpec{
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
				Tasks: []tektonv1.PipelineTask{
					{
						Name: "task-0",
						Params: []tektonv1.Param{
							{
								Name: "path",
								Value: tektonv1.ArrayOrString{
									Type:      tektonv1.ParamTypeString,
									StringVal: "test.yaml",
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
					},
				},
			}

			pipeline := cfg.ToPipelineSpec()
			Expect(pipeline).To(Equal(expected))
		})
	})
})
