// Package dex projects User and OAuth2 client configuration into Dex's
// Kubernetes-backed storage.
//
// Dex, when configured with `storage.type: kubernetes`, reads Password and
// OAuth2Client objects from the API server under the `dex.coreos.com/v1`
// group. The auth service mirrors our User custom resources into Password
// objects (Render/Sync) and ensures the gateway's OAuth2Client object exists
// (EnsureOAuth2Client). Dex consumes these at request time, so neither a
// ConfigMap rewrite nor a Dex restart is required to pick up changes.
package dex

import (
	"context"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	authv1alpha1 "gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/pkg/apis/auth/v1alpha1"
)

// Group and kinds of Dex's Kubernetes storage objects.
const (
	APIGroup         = "dex.coreos.com"
	APIVersion       = "v1"
	KindPassword     = "Password"
	KindOAuth2Client = "OAuth2Client"

	// ManagedByLabel marks objects owned by this controller so we can safely
	// prune stale entries without touching anything a human (or Dex itself)
	// put there.
	ManagedByLabel = "app.kubernetes.io/managed-by"
	// ManagedByValue is the value written to ManagedByLabel.
	ManagedByValue = "crossplane-ui-auth"
)

// PasswordGVK identifies Dex's Password resource.
var PasswordGVK = schema.GroupVersionKind{Group: APIGroup, Version: APIVersion, Kind: KindPassword}

// OAuth2ClientGVK identifies Dex's OAuth2Client resource.
var OAuth2ClientGVK = schema.GroupVersionKind{Group: APIGroup, Version: APIVersion, Kind: KindOAuth2Client}

// managedLabels are stamped on every object written by this package.
var managedLabels = map[string]string{
	ManagedByLabel:                ManagedByValue,
	"app.kubernetes.io/component": "dex-storage",
	"app.kubernetes.io/part-of":   "crossplane-ui",
}

// idToName mirrors Dex's storage/kubernetes name encoding:
// base32(lowercase alphabet, no padding) of the input bytes appended with the
// fnv-64 offset basis. Dex looks up Password and OAuth2Client objects by this
// exact name (see dexidp/dex storage/kubernetes/client.go idToName), so our
// writes must produce the same value.
func idToName(s string) string {
	enc := base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567")
	h := fnv.New64()
	return strings.TrimRight(enc.EncodeToString(h.Sum([]byte(s))), "=")
}

// Sync reconciles the Dex Password objects in namespace ns with the given
// User set: every enabled User with a non-empty password hash gets an upserted
// Password; any managed Password that no longer corresponds to a live User is
// deleted. The returned bool reports whether the cluster was mutated.
func Sync(ctx context.Context, c client.Client, ns string, users []authv1alpha1.User) (bool, error) {
	keep := make(map[string]struct{}, len(users))
	changed := false

	for i := range users {
		u := &users[i]
		if u.Spec.Disabled || u.Status.PasswordHash == "" || u.Status.UserID == "" {
			continue
		}
		email := strings.ToLower(u.Spec.Email)
		name := idToName(email)
		keep[name] = struct{}{}

		wrote, err := upsertPassword(ctx, c, ns, name, email, u)
		if err != nil {
			return changed, fmt.Errorf("upsert password %q: %w", u.Spec.Email, err)
		}
		if wrote {
			changed = true
		}
	}

	pruned, err := prunePasswords(ctx, c, ns, keep)
	if err != nil {
		return changed, err
	}
	if pruned {
		changed = true
	}
	return changed, nil
}

func upsertPassword(ctx context.Context, c client.Client, ns, name, email string, u *authv1alpha1.User) (bool, error) {
	existing := newPasswordObject()
	key := client.ObjectKey{Namespace: ns, Name: name}
	err := c.Get(ctx, key, existing)

	desired := renderPassword(ns, name, email, u)
	if apierrors.IsNotFound(err) {
		if err := c.Create(ctx, desired); err != nil {
			return false, fmt.Errorf("create: %w", err)
		}
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("get: %w", err)
	}

	if samePasswordPayload(existing, desired) {
		return false, nil
	}
	desired.SetResourceVersion(existing.GetResourceVersion())
	if err := c.Update(ctx, desired); err != nil {
		return false, fmt.Errorf("update: %w", err)
	}
	return true, nil
}

func renderPassword(ns, name, email string, u *authv1alpha1.User) *unstructured.Unstructured {
	obj := newPasswordObject()
	obj.SetNamespace(ns)
	obj.SetName(name)
	obj.SetLabels(copyLabels(managedLabels))

	// Dex's Password.Hash is `[]byte`; encoding/json serialises []byte as a
	// base64 string. We therefore store the (already-textual) bcrypt hash
	// wrapped in base64 so it round-trips through Dex's decoder unchanged.
	obj.Object["email"] = email
	obj.Object["hash"] = base64.StdEncoding.EncodeToString([]byte(u.Status.PasswordHash))
	obj.Object["username"] = u.Spec.Username
	obj.Object["userID"] = u.Status.UserID
	return obj
}

func samePasswordPayload(got, want *unstructured.Unstructured) bool {
	for _, k := range []string{"email", "hash", "username", "userID"} {
		if got.Object[k] != want.Object[k] {
			return false
		}
	}
	return true
}

func prunePasswords(ctx context.Context, c client.Client, ns string, keep map[string]struct{}) (bool, error) {
	list := newPasswordList()
	opts := []client.ListOption{
		client.InNamespace(ns),
		client.MatchingLabels(map[string]string{ManagedByLabel: ManagedByValue}),
	}
	if err := c.List(ctx, list, opts...); err != nil {
		return false, fmt.Errorf("list passwords: %w", err)
	}
	removed := false
	for i := range list.Items {
		item := &list.Items[i]
		if _, wanted := keep[item.GetName()]; wanted {
			continue
		}
		if err := c.Delete(ctx, item); err != nil && !apierrors.IsNotFound(err) {
			return removed, fmt.Errorf("delete password %q: %w", item.GetName(), err)
		}
		removed = true
	}
	return removed, nil
}

func newPasswordObject() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(PasswordGVK)
	return obj
}

func newPasswordList() *unstructured.UnstructuredList {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   APIGroup,
		Version: APIVersion,
		Kind:    KindPassword + "List",
	})
	return list
}

func copyLabels(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
