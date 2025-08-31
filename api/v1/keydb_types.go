/*
Copyright 2025.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KeydbSpec defines the desired state of Keydb.
type KeydbSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Keydb. Edit keydb_types.go to remove/update
	Image       string          `json:"image,omitempty"`
	Replicas    *int32          `json:"replicas,omitempty"`
	Replication ReplicationSpec `json:"replication,omitempty"`
	Persistence PersistenceSpec `json:"persistence,omitempty"`
	Password    string          `json:"password,omitempty"`
}

type ReplicationSpec struct {
	Enabled bool             `json:"enabled"`
	Mode    string           `json:"mode,omitempty"`
	Domain  []string         `json:"domain,omitempty"`
	Keydb   KeydbAddressSpec `json:"keydb,omitempty"`
	Port    int32            `json:"port,omitempty"`
}
type PersistenceSpec struct {
	Enabled          bool   `json:"enabled"`
	Size             string `json:"size,omitempty"`
	StorageClassName string `json:"storageClassName,omitempty"`
}
type KeydbAddressSpec struct {
	Namespace string `json:"host,omitempty"`
	Name      string `json:"name,omitempty"`
}

// KeydbStatus defines the observed state of Keydb.
type KeydbStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Keydb is the Schema for the keydbs API.
type Keydb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeydbSpec   `json:"spec,omitempty"`
	Status KeydbStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KeydbList contains a list of Keydb.
type KeydbList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Keydb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Keydb{}, &KeydbList{})
}
