/*
Copyright 2019 Andrew Rudoi.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-github/v28/github"
	yaml "gopkg.in/yaml.v2"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	knative "knative.dev/pkg/apis"

	alphav1 "github.com/rudoi/alaska/api/v1"
	"github.com/rudoi/alaska/pkg/alaska"
)

// RepoReconciler reconciles a Repo object
type RepoReconciler struct {
	client.Client
	GitHub *github.Client
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=alpha.alaska.rudeboy.io,resources=repos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=alpha.alaska.rudeboy.io,resources=repos/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelineresources;taskruns,verbs=get;list;watch;create;update;delete

func (r *RepoReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("repo", req.NamespacedName)

	repo := &alphav1.Repo{}
	if err := r.Get(ctx, req.NamespacedName, repo); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	patch := client.MergeFrom(repo.DeepCopyObject())

	defer func() {
		if err := r.Status().Patch(ctx, repo, patch); err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "error patching status")
		}
	}()

	if err := r.ensureTektonGitResource(ctx, repo); err != nil {
		log.Error(err, "unable to ensure PipelineResource")
		return ctrl.Result{}, nil
	}

	url, err := url.Parse(repo.Spec.URL)
	if err != nil {
		return ctrl.Result{}, err
	}

	owner := strings.Split(url.Path, "/")[1]
	repoName := strings.TrimSuffix(strings.Split(url.Path, "/")[2], ".git")

	branch, _, err := r.GitHub.Repositories.GetBranch(ctx, owner, repoName, repo.Spec.Branch)
	if err != nil {
		log.Error(err, "failed to get branch")
		return ctrl.Result{}, nil
	}

	sha := branch.GetCommit().GetSHA()[:7]
	content, _, _, err := r.GitHub.Repositories.GetContents(ctx, owner, repoName, "alaska.yaml", &github.RepositoryContentGetOptions{Ref: sha})
	if err != nil {
		log.Error(err, "unable to get config")
		return ctrl.Result{}, nil
	}

	decodedConfig, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		log.Error(err, "unable to decode config file")
		return ctrl.Result{}, nil
	}

	config := &alphav1.Config{}
	if err := yaml.Unmarshal(decodedConfig, config); err != nil {
		log.Error(err, "unable to unmarshal config")
		return ctrl.Result{}, nil
	}

	repo.Status.Config = config

	if err := r.ensurePipelineForRepo(ctx, repo, config); err != nil {
		log.Error(err, "unable to ensure pipeline for repo")
		return ctrl.Result{}, nil
	}

	log.V(4).Info("incoming config", "config", config)

	if repo.Status.CommitSHA != sha {
		log.Info("new commit detected", "branch", "master", "old", repo.Status.CommitSHA, "new", sha)
		repo.Status.CommitSHA = sha

		if err := r.patchGitResource(ctx, repo, sha); err != nil {
			return ctrl.Result{}, err
		}

		if err := alaska.TriggerPipeline(ctx, r.Client, repo, config, sha); err != nil {
			return ctrl.Result{}, err
		}
	}

	for i := range repo.Status.Runs {
		if !repo.Status.Runs[i].Completed {
			r.Log.Info("pizza time, run has NOT completed", "run", repo.Status.Runs[i])
			if err := r.updatePipelineRunStatus(ctx, repo, repo.Status.Runs[i], sha); err != nil {
				log.Error(err, "error waiting for pipeline to succeed")
				return ctrl.Result{}, nil
			}

			if repo.Status.Runs[i].Status == "Failed" {
				log.Info("pipeline failed, check the logs")
				repo.Status.Runs[i].Completed = true
				continue
			}

			if !repo.Status.Runs[i].Succeeded {
				log.Info("waiting for pipeline to succeed")
				return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
			}

			log.Info("pipeline for commit succeeded", "commit", sha)
			repo.Status.Runs[i].Completed = true
		}
	}

	return ctrl.Result{}, nil
}

func (r *RepoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alphav1.Repo{}).
		Complete(r)
}

func (r *RepoReconciler) ensureTektonGitResource(ctx context.Context, repo *alphav1.Repo) error {
	resource := &tektonv1.PipelineResource{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: repo.GetNamespace(), Name: repo.GetName()}, resource); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("")
			newResource := &tektonv1.PipelineResource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      repo.GetName(),
					Namespace: repo.GetNamespace(),
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: repo.APIVersion,
							Kind:       repo.Kind,
							Name:       repo.GetName(),
							UID:        repo.GetUID(),
						},
					},
				},
				Spec: tektonv1.PipelineResourceSpec{
					Type: tektonv1.PipelineResourceTypeGit,
					Params: []tektonv1.ResourceParam{
						{
							Name:  "url",
							Value: repo.Spec.URL,
						},
						{
							Name:  "revision",
							Value: repo.Spec.Branch,
						},
					},
				},
			}

			if err := r.Create(ctx, newResource); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	repo.Status.TektonRef = &corev1.ObjectReference{
		Name:      repo.GetName(),
		Namespace: repo.GetNamespace(),
	}

	return nil
}

func (r *RepoReconciler) patchGitResource(ctx context.Context, repo *alphav1.Repo, sha string) error {
	resource := &tektonv1.PipelineResource{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: repo.GetNamespace(), Name: repo.GetName()}, resource); err != nil {
		return err
	}

	patch := client.MergeFrom(resource.DeepCopyObject())

	resource.Spec.Params = []tektonv1.ResourceParam{
		{
			Name:  "url",
			Value: repo.Spec.URL,
		},
		{
			Name:  "revision",
			Value: sha,
		},
	}

	if err := r.Patch(ctx, resource, patch); err != nil {
		return err
	}

	return nil
}

func (r *RepoReconciler) updatePipelineRunStatus(ctx context.Context, repo *alphav1.Repo, runStatus *alphav1.PipelineStatus, sha string) error {
	query := types.NamespacedName{
		Namespace: runStatus.Ref.Namespace,
		Name:      runStatus.Ref.Name,
	}

	pipelineRun := &tektonv1.PipelineRun{}
	if err := r.Get(ctx, query, pipelineRun); err != nil {
		return err
	}

	runStatus.Ref = &corev1.ObjectReference{
		Name:       pipelineRun.GetName(),
		Namespace:  pipelineRun.GetNamespace(),
		Kind:       pipelineRun.Kind,
		APIVersion: pipelineRun.APIVersion,
		UID:        pipelineRun.GetUID(),
	}

	for _, condition := range pipelineRun.Status.Conditions {
		if condition.Type == knative.ConditionSucceeded {
			runStatus.Status = condition.Reason
			runStatus.Succeeded = condition.IsTrue()
		}
	}
	return nil
}

func (r *RepoReconciler) ensurePipelineForRepo(ctx context.Context, repo *alphav1.Repo, cfg *alphav1.Config) error {
	query := types.NamespacedName{
		Namespace: repo.GetNamespace(),
		Name:      repo.GetName(),
	}

	pipeline := &tektonv1.Pipeline{}
	err := r.Get(ctx, query, pipeline)
	if apierrors.IsNotFound(err) {
		// create
		pipeline = &tektonv1.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: repo.GetNamespace(),
				Name:      repo.GetName(),
			},
			Spec: cfg.ToPipelineSpec(),
		}

		return r.Create(ctx, pipeline)
	}

	if err != nil {
		return err
	}

	// update
	pipeline.Spec = cfg.ToPipelineSpec()
	return r.Update(ctx, pipeline)
}
