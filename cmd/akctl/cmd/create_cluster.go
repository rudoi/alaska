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
		if err := RunClusterCreate(co); err != nil {
			klog.Exit(err)
		}
	},
}

func RunClusterCreate(co *ClusterOptions) error {
	return nil
}

func init() {
	createCmd.AddCommand(createClusterCmd)
}
