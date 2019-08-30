package cmd

import (
	"context"
	"flag"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"k8s.io/klog"
)

type ServiceAccountOptions struct {
	Name             string
	TargetKubeconfig string
	TargetNamespace  string
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = corev1.AddToScheme(scheme)
	_ = tektonv1.AddToScheme(scheme)
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

func RunServiceAccountCreate(sao *ServiceAccountOptions) error {
	_ = flag.Set("kubeconfig", sao.TargetKubeconfig)
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	client, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil
	}

	ctx := context.Background()

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sao.Name,
			Namespace: sao.TargetNamespace,
		},
	}
	if err := client.Create(ctx, sa); err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	if err := client.Get(ctx, types.NamespacedName{Namespace: sao.TargetNamespace, Name: sao.Name}, sa); err != nil {
		return err
	}

	secret := &corev1.Secret{}
	if err := client.Get(ctx, types.NamespacedName{Namespace: sao.TargetNamespace, Name: sa.Secrets[0].Name}, secret); err != nil {
		return err
	}

	resource := &tektonv1.PipelineResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sao.Name,
			Namespace: sao.TargetNamespace,
		},
		Spec: tektonv1.PipelineResourceSpec{
			Type: tektonv1.PipelineResourceTypeCluster,
			Params: []tektonv1.ResourceParam{
				{
					Name:  "name",
					Value: sao.Name,
				},
				{
					Name:  "username",
					Value: sao.Name,
				},
				{
					Name:  "url",
					Value: cfg.Host,
				},
				{
					Name:  "cadata",
					Value: string(secret.Data["ca.crt"]),
				},
				{
					Name:  "token",
					Value: string(secret.Data["token"]),
				},
			},
		},
	}

	return client.Create(ctx, resource)
}

func init() {
	home := homedir.HomeDir()
	// required
	createServiceAccountCmd.Flags().StringVarP(&sao.Name, "name", "", "", "name for set of Kubernetes credentials - required")
	_ = createServiceAccountCmd.MarkFlagRequired("name")

	// optional
	createServiceAccountCmd.Flags().StringVarP(&sao.TargetKubeconfig, "target-kubeconfig", "s", filepath.Join(home, ".kube", "config"), "kubeconfig to use for creating serviceaccount")
	createServiceAccountCmd.Flags().StringVarP(&sao.TargetNamespace, "target-namespace", "", "default", "namespace for new serviceaccount")

	createCmd.AddCommand(createServiceAccountCmd)
}
