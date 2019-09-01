package alaska

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			pipeline := cfg.ToPipelineSpec()
			Expect(pipeline.Tasks).To(HaveLen(1))
		})
	})
})
