package cmd

import (
	"context"

	alphav1 "github.com/rudoi/alaska/api/v1"
	"github.com/rudoi/alaska/pkg/alaska"
	"github.com/spf13/cobra"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type RetryOptions struct {
	Namespace string
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = corev1.AddToScheme(scheme)
	_ = alphav1.AddToScheme(scheme)
	_ = tektonv1.AddToScheme(scheme)
}

var ro = &RetryOptions{}
var retryCmd = &cobra.Command{
	Use:   "retry",
	Short: "retry an Alaska pipeline",
	Long:  "trigger another build for a specific Repo",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunRetry(args[0], ro); err != nil {
			klog.Exit(err)
		}
	},
}

func RunRetry(name string, ro *RetryOptions) error {
	ctx := context.Background()
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil
	}

	query := types.NamespacedName{
		Namespace: ro.Namespace,
		Name:      name,
	}

	repo := &alphav1.Repo{}
	if err := c.Get(ctx, query, repo); err != nil {
		return err
	}

	patch := client.MergeFrom(repo.DeepCopyObject())

	if err := alaska.TriggerPipeline(ctx, c, repo, repo.Status.Config, repo.Status.CommitSHA); err != nil {
		return err
	}

	if err := c.Status().Patch(ctx, repo, patch); err != nil {
		return err
	}

	return nil
}

func init() {
	// optional
	createServiceAccountCmd.Flags().StringVarP(&ro.Namespace, "namespace", "", "default", "namespace repo is in")

	rootCmd.AddCommand(retryCmd)
}
