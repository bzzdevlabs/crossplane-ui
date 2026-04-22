package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GroupSpec is the desired state of a local group.
type GroupSpec struct {
	// DisplayName is a human-friendly label shown in the UI.
	// +optional
	// +kubebuilder:validation:MaxLength=128
	DisplayName string `json:"displayName,omitempty"`

	// Description is a free-form description of the group's purpose.
	// +optional
	// +kubebuilder:validation:MaxLength=512
	Description string `json:"description,omitempty"`
}

// GroupStatus is the observed state of a Group.
type GroupStatus struct {
	// Members lists the usernames currently pointing at this group through
	// their .spec.groups. It is maintained by the User controller and is
	// purely informational.
	// +optional
	Members []string `json:"members,omitempty"`

	// ObservedGeneration is the .metadata.generation the controller last
	// reconciled successfully.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions reports the group's reconcile state.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=xpuigroup
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Display Name",type=string,JSONPath=`.spec.displayName`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Group represents a named collection of Users. Group membership is carried
// by User.spec.groups; the controller mirrors the set on this status for
// quick inspection with `kubectl get groups`.
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec,omitempty"`
	Status GroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupList is a list of Group.
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Group{}, &GroupList{})
}
