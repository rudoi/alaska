# Alaska :snowboarder: - GitOps CI/CD for Kubernetes :ship:

Alaska is a Kubernetes-native CI/CD orchestrator backed by [Tekton](https://github.com/tektoncd/pipeline). It responds to changes on a specified branch of a GitHub repo and applies Kubernetes manifests contained within.

The goal is to create a Kubernetes-targeted deployment platform that's easy to administer as an operator _and_ easy to consume as a developer. Oh, also to have fun along the way! :pizza:

## What does it do?

The Alaska controller watches GitHub repositories for changes. Each repo is paired with a set of Kubernetes credentials. When a change is detected, the controller:

1. Grabs an `alaska.yaml` config file from the repo
2. Creates/updates a Tekton Pipeline based on the config. The Pipeline applies Kubernetes YAML to the cluster associated with the repo.
3. Triggers the Pipeline
4. Polls until a finite status is returned

The controller operates over Kubernetes custom resources called Repos. They look like this:

```yaml
apiVersion: alpha.alaska.rudeboy.io/v1
kind: Repo
metadata:
  name: repo-sample
spec:
  url: https://github.com/rudoi/alaska-test.git
  branch: master
  cluster: pizza
```

This watches for changes in the `rudoi/alaska-test` GitHub repository on the `master` branch. Manifests specified in the `alaska.yaml` in the root of that repository are applied to the `pizza` Kubernetes cluster. The controller expects there to be a Tekton [PipelineResource](https://github.com/tektoncd/pipeline/blob/master/docs/resources.md#cluster-resource) of type `cluster` in the same namespace as the `Repo` object.

## Configuration

Here's an annotated example `alaska.yaml`:

```yaml
manifests:
  # kubectl apply -f configmap.yaml
  - path: configmap.yaml

  # helm upgrade --install charts/test-chart test-chart
  - path: charts/test-chart
    type: helm
```

The controller, upon seeing new commits to the repo, will perform the actions in the comments above.

**note**: the Tekton ClusterTask that invokes `helm` is hard-coded to use [Helm 3.0.0-beta.2](https://github.com/helm/helm/releases/tag/v3.0.0-beta.2).

## Getting Started

Nope, coming soon! :sweat_smile:

## Features

### controller

- [ ] configurable target namespace
- [ ] remote helm charts with local values.yaml
- [ ] multi-cluster deploys
- [ ] ConfigMap configuration option
- [ ] configurable ordering (apply `crds/` then apply `manifests/`, etc)
- [ ] define `kustomize` executor
- [x] define `helm` executor
- [x] define `kubectl` executor
- [x] specify multiple paths for manifests
- [x] fetch `alaska.yaml` from repo for additional configuration
- [x] trigger pipeline on changes
- [x] specify branch to watch
- [x] watch repo for changes

stretch goals:

- [ ] object-granular status reporting
- [ ] pull request actions
- [ ] define generic executor (use image x, run command y, etc)

### `akctl` CLI

- [ ] create Repo with any required credentials (in single command)
- [x] manually retry latest build for a repo
- [x] create serviceaccount and generate Kubernetes credentials for Alaska controller to use (in single command)

## Should I use this?

Not yet!
