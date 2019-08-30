package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

type ServiceAccountOptions struct {
	Name string
}

var sao = &ServiceAccountOptions{}

var createServiceAccountCmd = &cobra.Command{
	Use:   "serviceaccount",
	Short: "create a K8S ServiceAccount for use with Alaska",
	Long:  "create a Kubernetes ServiceAccount and make its credentials available to Alaska",
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunServiceAccountCreate(sao); err != nil {
			klog.Exit(err)
		}
	},
}

func RunServiceAccountCreate(co *ServiceAccountOptions) error {
	return nil
}

func init() {
	createServiceAccountCmd.Flags().StringVarP(&sao.Name, "name", "", "", "name for set of Kubernetes credentials - required")
	_ = createServiceAccountCmd.MarkFlagRequired("name")

	createCmd.AddCommand(createServiceAccountCmd)
}
