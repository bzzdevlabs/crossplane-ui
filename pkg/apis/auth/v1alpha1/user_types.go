package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UserSpec is the desired state of a local user.
type UserSpec struct {
	// Email is the user's RFC 5322 e-mail address. It is used as the stable
	// identifier passed to Dex (sub claim) and as the login identity in the
	// password DB connector.
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:Pattern=`^[^\s@]+@[^\s@]+\.[^\s@]+$`
	Email string `json:"email"`

	// Username is the human-friendly login name surfaced in the UI. It must be
	// unique across all User objects; the controller reports conflicts on the
	// status of the offending objects.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._-]+$`
	Username string `json:"username"`

	// PasswordSecretRef references a Secret in the auth service's namespace
	// that carries a "password" key (plain text). The auth controller reads
	// the Secret, bcrypts the value and populates Status.PasswordHash, then
	// clears the Secret's password key.
	//
	// If PasswordHash is set directly in the Status by an administrator (for
	// imports from an existing system) PasswordSecretRef may be omitted.
	// +optional
	PasswordSecretRef *SecretReference `json:"passwordSecretRef,omitempty"`

	// Groups lists the Group names the user belongs to. Groups are resolved
	// against Group objects by name. Unknown groups are ignored (the
	// controller sets a Degraded condition).
	// +optional
	Groups []string `json:"groups,omitempty"`

	// Disabled hides the user from Dex's staticPasswords list without deleting
	// the object. Useful for temporary lockouts.
	// +optional
	Disabled bool `json:"disabled,omitempty"`
}

// SecretReference points to a Secret in the auth service's namespace.
type SecretReference struct {
	// Name of the Secret.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// UserStatus is the observed state of a User.
type UserStatus struct {
	// PasswordHash is the bcrypt hash of the user's password. It is set by the
	// controller after consuming PasswordSecretRef, and is the single source
	// of truth projected into the Dex configuration.
	// +optional
	PasswordHash string `json:"passwordHash,omitempty"`

	// UserID is a stable, randomly generated identifier the controller emits
	// on creation. It is fed to Dex as the userID of the staticPasswords
	// entry and therefore becomes the OIDC sub claim.
	// +optional
	UserID string `json:"userID,omitempty"`

	// ObservedGeneration is the .metadata.generation the controller last
	// reconciled successfully.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions reports the user's reconcile state.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// Condition types used on User and Group status.
const (
	// ConditionReady signals that the object has been fully projected into
	// Dex (or, for a Group, that its members are reconciled).
	ConditionReady = "Ready"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=xpuiuser
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Username",type=string,JSONPath=`.spec.username`
// +kubebuilder:printcolumn:name="Email",type=string,JSONPath=`.spec.email`
// +kubebuilder:printcolumn:name="Disabled",type=boolean,JSONPath=`.spec.disabled`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// User represents a local account authenticating through Dex's password DB.
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserList is a list of User.
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}
