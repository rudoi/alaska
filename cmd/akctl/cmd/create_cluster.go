package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

type ClusterOptions struct {
	Name string
}

var co = &ClusterOptions{}

var createClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "save Kubernetes authentication information for use with Alaska",
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunClusterCreate(co); err != nil {
			klog.Exit(err)
		}
	},
}

func RunClusterCreate(co *ClusterOptions) error {
	return nil
}

func init() {
	createClusterCmd.Flags().StringVarP(&co.Name, "name", "", "", "name for set of Kubernetes credentials - required")
	_ = createClusterCmd.MarkFlagRequired("name")

	createCmd.AddCommand(createClusterCmd)
}
