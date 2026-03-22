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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KeydbSpec defines the desired state of Keydb.
type KeydbSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Image defines the container image to use.
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image,omitempty"`
	// +kubebuilder:validation:Minimum=1
	Replicas    *int32          `json:"replicas,omitempty"`
	Replication ReplicationSpec `json:"replication,omitempty"`
	Persistence PersistenceSpec `json:"persistence,omitempty"`
	// PasswordSecret is a reference to the secret containing the password for KeyDB.
	// If not specified, a random password will be generated and stored in a new Secret.
	// +optional
	PasswordSecret *corev1.SecretKeySelector `json:"passwordSecret,omitempty"`
	// Resources defines the resource requests and limits for KeyDB pods
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Metrics enables exposing Prometheus metrics via an exporter sidecar
	// +optional
	Metrics MetricsSpec `json:"metrics,omitempty"`
}

type MetricsSpec struct {
	Enabled bool   `json:"enabled"`
	Image   string `json:"image,omitempty"`
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
	// Phase represents the current phase of the KeyDB cluster
	Phase string `json:"phase,omitempty"`
	// Conditions represent the latest available observations of the KeyDB cluster's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// Replicas represents the current replica status
	Replicas ReplicaStatus `json:"replicas,omitempty"`
	// ReadyReplicas is the number of KeyDB pods that are ready
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// CurrentReplicas is the current number of replicas
	CurrentReplicas int32 `json:"currentReplicas,omitempty"`
	// ObservedGeneration reflects the generation of the most recently observed KeyDB
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastUpdateTime is the last time the status was updated
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
}

// ReplicaStatus represents the status of individual replicas
type ReplicaStatus struct {
	Ready    []string `json:"ready,omitempty"`
	NotReady []string `json:"notReady,omitempty"`
	Failed   []string `json:"failed,omitempty"`
}

// Condition types
const (
	ConditionTypeReady       = "Ready"
	ConditionTypeDegraded    = "Degraded"
	ConditionTypeProgressing = "Progressing"
	ConditionTypeReconciled  = "Reconciled"
)

// Condition reasons
const (
	ReasonReconcileSuccess     = "ReconcileSuccess"
	ReasonReconcileError       = "ReconcileError"
	ReasonScalingUp            = "ScalingUp"
	ReasonScalingDown          = "ScalingDown"
	ReasonAllReplicasReady     = "AllReplicasReady"
	ReasonSomeReplicasNotReady = "SomeReplicasNotReady"
)

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
