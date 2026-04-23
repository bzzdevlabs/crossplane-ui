package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ConnectorType enumerates the Dex connector kinds the UI supports natively.
// Dex itself supports more; these are the five wired up by M10.
// +kubebuilder:validation:Enum=ldap;saml;github;google;oidc
type ConnectorType string

// Supported connector types.
const (
	ConnectorTypeLDAP   ConnectorType = "ldap"
	ConnectorTypeSAML   ConnectorType = "saml"
	ConnectorTypeGitHub ConnectorType = "github"
	ConnectorTypeGoogle ConnectorType = "google"
	ConnectorTypeOIDC   ConnectorType = "oidc"
)

// ConnectorSecretInjection declares a value the reconciler must splice into
// the Dex config at a specific JSON path before writing the Connector
// storage object. Secrets therefore never appear in the CR itself.
type ConnectorSecretInjection struct {
	// Path is a dot-separated JSON path into .spec.config, e.g.
	// "clientSecret" or "bindPW" or "rootCAData".
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path"`

	// SecretRef points to the Secret that carries the value.
	SecretRef ConnectorSecretRef `json:"secretRef"`
}

// ConnectorSecretRef references a Secret key in the auth service's namespace.
type ConnectorSecretRef struct {
	// Name of the Secret.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Key within the Secret's data map.
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`
}

// ConnectorSpec is the desired state of a Dex connector.
type ConnectorSpec struct {
	// ID is the stable Dex connector id. It becomes part of the callback
	// path (/callback/{id}) and therefore must be URL-safe and stable across
	// edits (changing it invalidates existing IdP trust).
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	ID string `json:"id"`

	// Type is the connector kind. Must match a connector Dex knows how to
	// instantiate.
	Type ConnectorType `json:"type"`

	// Name is the human-readable label rendered on the Dex login chooser.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	Name string `json:"name"`

	// Config is the raw Dex connector config. Its schema depends on .type —
	// see docs/authentication.md for per-provider examples. Stored as an
	// opaque JSON object; the controller marshals it unchanged (after secret
	// injection) into Dex's storage.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Config runtime.RawExtension `json:"config"`

	// SecretRefs lists values spliced into .spec.config at project-time so
	// plaintext secrets (client secret, LDAP bind password, certificates)
	// live only in Secrets. The path is a dot-separated JSON path within
	// .spec.config.
	// +optional
	SecretRefs []ConnectorSecretInjection `json:"secretRefs,omitempty"`

	// Disabled hides the connector from Dex without deleting the CR.
	// +optional
	Disabled bool `json:"disabled,omitempty"`
}

// ConnectorStatus is the observed state of a Connector.
type ConnectorStatus struct {
	// ObservedGeneration is the .metadata.generation the controller last
	// reconciled successfully.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions reports the connector's reconcile state.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=xpuiconn
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Disabled",type=boolean,JSONPath=`.spec.disabled`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Connector represents a Dex upstream identity provider (LDAP, SAML, GitHub,
// Google, OIDC). The auth controller projects Connectors into Dex's
// `dex.coreos.com/v1` Connector storage after inlining any secret-backed
// fields.
type Connector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectorSpec   `json:"spec,omitempty"`
	Status ConnectorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConnectorList is a list of Connector.
type ConnectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Connector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Connector{}, &ConnectorList{})
}
