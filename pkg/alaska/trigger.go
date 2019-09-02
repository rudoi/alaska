package alaska

import (
	"context"
	"fmt"

	alphav1 "github.com/rudoi/alaska/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

func TriggerPipeline(ctx context.Context, c client.Client, repo *alphav1.Repo, config *alphav1.Config, sha string) error {
	pipelineRun := &tektonv1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-", repo.GetName(), sha),
			Namespace:    repo.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: alphav1.GroupVersion.Version,
					Kind:       "Repo",
					Name:       repo.GetName(),
					UID:        repo.GetUID(),
				},
			},
		},
		Spec: tektonv1.PipelineRunSpec{
			Resources: []tektonv1.PipelineResourceBinding{
				{
					Name: "repo",
					ResourceRef: tektonv1.PipelineResourceRef{
						Name: repo.GetName(),
					},
				},
				{
					Name: "cluster",
					ResourceRef: tektonv1.PipelineResourceRef{
						Name: repo.Spec.Cluster,
					},
				},
			},
			PipelineRef: tektonv1.PipelineRef{
				Name: repo.GetName(),
			},
		},
	}

	if err := c.Create(ctx, pipelineRun); err != nil {
		return err
	}

	status := &alphav1.PipelineStatus{
		Ref: &corev1.ObjectReference{
			Name:       pipelineRun.GetName(),
			Namespace:  pipelineRun.GetNamespace(),
			Kind:       pipelineRun.Kind,
			APIVersion: pipelineRun.APIVersion,
			UID:        pipelineRun.GetUID(),
		},
	}

	// put latest in front, limit to 5 total
	if len(repo.Status.Runs) >= 4 {
		repo.Status.Runs = append([]*alphav1.PipelineStatus{status}, repo.Status.Runs[:3]...)
	} else {
		repo.Status.Runs = append([]*alphav1.PipelineStatus{status}, repo.Status.Runs...)
	}
	return nil
}
