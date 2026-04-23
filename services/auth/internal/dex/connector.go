package dex

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	authv1alpha1 "github.com/bzzdevlabs/crossplane-ui/pkg/apis/auth/v1alpha1"
)

// KindConnector is the Dex storage kind for upstream identity providers.
const KindConnector = "Connector"

// ConnectorGVK identifies Dex's Connector storage resource.
var ConnectorGVK = schema.GroupVersionKind{Group: APIGroup, Version: APIVersion, Kind: KindConnector}

// SecretResolver reads a value out of a Secret in ns by name and key.
type SecretResolver func(ctx context.Context, ns, name, key string) (string, error)

// KubeSecretResolver returns a SecretResolver backed by the given client.
func KubeSecretResolver(c client.Client) SecretResolver {
	return func(ctx context.Context, ns, name, key string) (string, error) {
		var sec corev1.Secret
		if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &sec); err != nil {
			return "", fmt.Errorf("get secret %q: %w", name, err)
		}
		raw, ok := sec.Data[key]
		if !ok {
			return "", fmt.Errorf("secret %q missing key %q", name, key)
		}
		return string(raw), nil
	}
}

// SyncConnectors reconciles Dex Connector objects in namespace ns from the
// given Connector CR set. Secrets referenced by each CR are read via resolve
// and spliced into the projected config.
func SyncConnectors(
	ctx context.Context,
	c client.Client,
	resolve SecretResolver,
	ns string,
	conns []authv1alpha1.Connector,
) (bool, error) {
	keep := make(map[string]struct{}, len(conns))
	changed := false

	for i := range conns {
		conn := &conns[i]
		if conn.Spec.Disabled {
			continue
		}
		name := idToName(conn.Spec.ID)
		keep[name] = struct{}{}

		wrote, err := upsertConnector(ctx, c, resolve, ns, name, conn)
		if err != nil {
			return changed, fmt.Errorf("upsert connector %q: %w", conn.Spec.ID, err)
		}
		if wrote {
			changed = true
		}
	}

	pruned, err := pruneConnectors(ctx, c, ns, keep)
	if err != nil {
		return changed, err
	}
	if pruned {
		changed = true
	}
	return changed, nil
}

func upsertConnector(
	ctx context.Context,
	c client.Client,
	resolve SecretResolver,
	ns, name string,
	conn *authv1alpha1.Connector,
) (bool, error) {
	desired, err := renderConnector(ctx, resolve, ns, name, conn)
	if err != nil {
		return false, err
	}

	existing := newConnectorObject()
	err = c.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, existing)
	if apierrors.IsNotFound(err) {
		if err := c.Create(ctx, desired); err != nil {
			return false, fmt.Errorf("create: %w", err)
		}
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("get: %w", err)
	}

	if sameConnectorPayload(existing, desired) {
		return false, nil
	}
	desired.SetResourceVersion(existing.GetResourceVersion())
	if err := c.Update(ctx, desired); err != nil {
		return false, fmt.Errorf("update: %w", err)
	}
	return true, nil
}

func renderConnector(
	ctx context.Context,
	resolve SecretResolver,
	ns, name string,
	conn *authv1alpha1.Connector,
) (*unstructured.Unstructured, error) {
	cfg := map[string]any{}
	if raw := conn.Spec.Config.Raw; len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, fmt.Errorf("decode config: %w", err)
		}
	}

	for _, inj := range conn.Spec.SecretRefs {
		if resolve == nil {
			return nil, fmt.Errorf("secret resolver is nil but connector %q has secretRefs", conn.Spec.ID)
		}
		val, err := resolve(ctx, ns, inj.SecretRef.Name, inj.SecretRef.Key)
		if err != nil {
			return nil, fmt.Errorf("resolve %s: %w", inj.Path, err)
		}
		if err := setAtPath(cfg, inj.Path, val); err != nil {
			return nil, fmt.Errorf("inject %s: %w", inj.Path, err)
		}
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	obj := newConnectorObject()
	obj.SetNamespace(ns)
	obj.SetName(name)
	obj.SetLabels(copyLabels(managedLabels))
	obj.Object["id"] = conn.Spec.ID
	obj.Object["type"] = string(conn.Spec.Type)
	obj.Object["name"] = conn.Spec.Name
	// Dex's Connector.Config is []byte; JSON serialises []byte as a base64
	// string, so store the JSON-encoded config wrapped in base64.
	obj.Object["config"] = base64.StdEncoding.EncodeToString(cfgBytes)
	return obj, nil
}

func sameConnectorPayload(got, want *unstructured.Unstructured) bool {
	for _, k := range []string{"id", "type", "name", "config"} {
		if got.Object[k] != want.Object[k] {
			return false
		}
	}
	return true
}

func pruneConnectors(ctx context.Context, c client.Client, ns string, keep map[string]struct{}) (bool, error) {
	list := newConnectorList()
	opts := []client.ListOption{
		client.InNamespace(ns),
		client.MatchingLabels(map[string]string{ManagedByLabel: ManagedByValue}),
	}
	if err := c.List(ctx, list, opts...); err != nil {
		return false, fmt.Errorf("list connectors: %w", err)
	}
	removed := false
	for i := range list.Items {
		item := &list.Items[i]
		if _, wanted := keep[item.GetName()]; wanted {
			continue
		}
		if err := c.Delete(ctx, item); err != nil && !apierrors.IsNotFound(err) {
			return removed, fmt.Errorf("delete connector %q: %w", item.GetName(), err)
		}
		removed = true
	}
	return removed, nil
}

func newConnectorObject() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(ConnectorGVK)
	return obj
}

func newConnectorList() *unstructured.UnstructuredList {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   APIGroup,
		Version: APIVersion,
		Kind:    KindConnector + "List",
	})
	return list
}

// setAtPath writes val into m at a dot-separated JSON path, creating missing
// intermediate maps. Fails if an intermediate path segment is already a
// non-map value.
func setAtPath(m map[string]any, path, val string) error {
	if path == "" {
		return errors.New("empty path")
	}
	parts := strings.Split(path, ".")
	cur := m
	for i, p := range parts {
		if p == "" {
			return fmt.Errorf("empty segment at %d", i)
		}
		if i == len(parts)-1 {
			cur[p] = val
			return nil
		}
		next, ok := cur[p]
		if !ok {
			nm := map[string]any{}
			cur[p] = nm
			cur = nm
			continue
		}
		nm, ok := next.(map[string]any)
		if !ok {
			return fmt.Errorf("segment %q is not an object", p)
		}
		cur = nm
	}
	return nil
}
