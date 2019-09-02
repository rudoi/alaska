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

package v1

import (
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RepoSpec defines the desired state of Repo
type RepoSpec struct {
	URL     string `json:"url"`
	Branch  string `json:"branch"`
	Cluster string `json:"cluster"`
}

type PipelineStatus struct {
	Completed bool                    `json:"completed,omitempty"`
	Ref       *corev1.ObjectReference `json:"ref,omitempty"`
	Status    string                  `json:"status,omitempty"`
	Succeeded bool                    `json:"succeeded,omitempty"`
}

// RepoStatus defines the observed state of Repo
type RepoStatus struct {
	CommitSHA string                  `json:"commitSHA,omitempty"`
	Config    *Config                 `json:"config,omitempty"`
	TektonRef *corev1.ObjectReference `json:"tektonRef,omitempty"`
	Runs      []*PipelineStatus       `json:"runs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=repos,shortName=rp
// +kubebuilder:subresource:status

// Repo is the Schema for the repos API
type Repo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepoSpec   `json:"spec,omitempty"`
	Status RepoStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RepoList contains a list of Repo
type RepoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Repo{}, &RepoList{})
}
