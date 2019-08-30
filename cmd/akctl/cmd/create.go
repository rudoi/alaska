package cmd

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create an Alaska resource",
	Long:  "create an Alaska resource, such as a Repo or a Kubernetes cluster Secret",
}

func init() {
	rootCmd.AddCommand(createCmd)
}
