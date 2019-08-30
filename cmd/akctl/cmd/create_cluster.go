package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

type ClusterOptions struct {
}

var co = &ClusterOptions{}

var createClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "save Kubernetes authentication information for use with Alaska",
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunCreate(co); err != nil {
			klog.Exit(err)
		}
	},
}

func RunCreate(co *ClusterOptions) error {
	return nil
}

func init() {
	createCmd.AddCommand(createClusterCmd)
}
