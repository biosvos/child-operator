/*
Copyright 2024.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MineSpec defines the desired state of Mine
type MineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Mine. Edit mine_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// MineStatus defines the observed state of Mine
type MineStatus struct {
	ChildResourceName string `json:"child_resource_name,omitempty"`
	CopyChildStatus   string `json:"copy_child_status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Mine is the Schema for the mines API
type Mine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MineSpec   `json:"spec,omitempty"`
	Status MineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MineList contains a list of Mine
type MineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mine `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&Mine{}, &MineList{})
}
