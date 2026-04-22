// Package v1alpha1 contains the auth.crossplane-ui.io/v1alpha1 API group.
//
// The group declares two cluster-scoped kinds: User and Group. They back the
// "local users" identity store consumed by Dex through the Dex <-> auth
// controller sync loop. See docs/adr/0004-dex-as-auth-surface.md.
//
// +kubebuilder:object:generate=true
// +groupName=auth.crossplane-ui.io
package v1alpha1
