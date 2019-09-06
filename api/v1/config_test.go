package v1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

var _ = Describe("Config.ToPipelineSpec tests", func() {
	var (
		cfg      *Config
		expected tektonv1.PipelineSpec
	)

	BeforeEach(func() {
		expected = tektonv1.PipelineSpec{
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
						Name: ExecutorDefault.toTaskName(),
						Kind: tektonv1.ClusterTaskKind,
					},
				},
			},
		}
	})

	Context("given a config with a valid kubectl manifest path", func() {
		BeforeEach(func() {
			cfg = &Config{
				Manifests: []*ManifestOptions{&ManifestOptions{Path: "test.yaml"}},
			}
		})

		It("should return a pipeline spec with one task", func() {
			pipeline := cfg.ToPipelineSpec()
			Expect(pipeline).To(Equal(expected))
		})
	})

	Context("given a config with a valid helm chart path", func() {
		BeforeEach(func() {
			cfg = &Config{
				Manifests: []*ManifestOptions{
					{
						Path: "path/to/chart",
						Type: ExecutorHelm,
					},
				},
			}
		})

		It("should return a pipeline spec with one task", func() {
			expected.Tasks[0].TaskRef.Name = ExecutorHelm.toTaskName()
			expected.Tasks[0].Params = []tektonv1.Param{
				{
					Name: "path",
					Value: tektonv1.ArrayOrString{
						Type:      tektonv1.ParamTypeString,
						StringVal: "path/to/chart",
					},
				},
				{
					Name: "release",
					Value: tektonv1.ArrayOrString{
						Type:      tektonv1.ParamTypeString,
						StringVal: "chart",
					},
				},
			}

			pipeline := cfg.ToPipelineSpec()
			Expect(pipeline).To(Equal(expected))
		})
	})

	Context("given multiple paths and sequential execution", func() {
		BeforeEach(func() {
			cfg = &Config{
				Manifests: []*ManifestOptions{{Path: "test-0.yaml"}, {Path: "test-1.yaml"}},
				Strategy:  StrategySequential,
			}
		})

		It("should create a pipeline that executes the paths in order", func() {
			pipeline := cfg.ToPipelineSpec()
			Expect(pipeline.Tasks[1].RunAfter).To(Equal([]string{"task-0"}))
		})
	})
})
