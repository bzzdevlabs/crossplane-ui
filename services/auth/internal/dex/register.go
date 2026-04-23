package dex

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// AddToScheme registers Dex's Kubernetes storage kinds (Password and
// OAuth2Client) on a runtime scheme as unstructured types. We don't carry Go
// structs for them because only a handful of fields are written and the
// unstructured encoder sidesteps mirroring Dex's private type definitions.
//
// Registering keeps controller-runtime's fake client and RESTMapper happy
// without having to race-add the types on first use.
func AddToScheme(scheme *runtime.Scheme) {
	passwordGV := PasswordGVK.GroupVersion()
	scheme.AddKnownTypeWithName(PasswordGVK, &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(passwordGV.WithKind(KindPassword+"List"), &unstructured.UnstructuredList{})

	clientGV := OAuth2ClientGVK.GroupVersion()
	scheme.AddKnownTypeWithName(OAuth2ClientGVK, &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(clientGV.WithKind(KindOAuth2Client+"List"), &unstructured.UnstructuredList{})

	connectorGV := ConnectorGVK.GroupVersion()
	scheme.AddKnownTypeWithName(ConnectorGVK, &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(connectorGV.WithKind(KindConnector+"List"), &unstructured.UnstructuredList{})
}
