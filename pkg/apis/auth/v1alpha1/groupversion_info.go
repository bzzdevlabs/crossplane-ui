package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// GroupVersion identifies the auth.crossplane-ui.io/v1alpha1 API.
var GroupVersion = schema.GroupVersion{Group: "auth.crossplane-ui.io", Version: "v1alpha1"}

// SchemeBuilder registers the group's types with a runtime scheme.
var SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

// AddToScheme adds the types in this API group to the given scheme.
var AddToScheme = SchemeBuilder.AddToScheme
