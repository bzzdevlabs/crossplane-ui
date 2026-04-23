package api

import (
	"context"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

// CrossplaneFactory builds impersonated dynamic + discovery clients. The
// gateway's kube.ClientFactory satisfies it; tests can plug in fakes.
type CrossplaneFactory interface {
	Dynamic(user string, groups []string) (dynamic.Interface, error)
	Discovery(user string, groups []string) (discovery.DiscoveryInterface, error)
}

// CrossplaneResource is the projected shape of a single Crossplane-related
// resource as surfaced on the home dashboard. We deliberately do not emit
// the full unstructured object; only what the UI needs to render a tile.
type CrossplaneResource struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	// Resource is the plural/lowercase resource name (e.g. "providers"),
	// needed by the UI to target this object through the generic resource
	// endpoint. We surface it alongside Kind to avoid a second discovery
	// round-trip from the browser.
	Resource          string    `json:"resource"`
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace,omitempty"`
	Ready             string    `json:"ready"`
	Synced            string    `json:"synced"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
}

// CrossplaneGroup bundles resources by dashboard category.
type CrossplaneGroup struct {
	Category string               `json:"category"`
	Items    []CrossplaneResource `json:"items"`
	// Error is non-empty when the category could not be listed (RBAC, CRD not
	// installed, transient API failure). The UI renders the category tile
	// with the error message so operators can diagnose partial failures.
	Error string `json:"error,omitempty"`
}

// CrossplaneSummary is the envelope returned by /api/v1/crossplane/resources.
type CrossplaneSummary struct {
	Groups []CrossplaneGroup `json:"groups"`
}

// knownTopLevel describes a Crossplane resource we list directly by GVR.
// Managed and composite resources are discovered from the API server rather
// than hard-coded because their GVRs live in provider packages and XRDs.
type knownTopLevel struct {
	Category string
	GVR      schema.GroupVersionResource
}

// topLevelGVRs are the Crossplane resources every installation is expected
// to host. They are listed by fixed GVR; if a category's CRD is missing the
// handler returns an empty list for it (not an error).
var topLevelGVRs = []knownTopLevel{
	{Category: "composition", GVR: schema.GroupVersionResource{
		Group: "apiextensions.crossplane.io", Version: "v1", Resource: "compositions",
	}},
	{Category: "provider", GVR: schema.GroupVersionResource{
		Group: "pkg.crossplane.io", Version: "v1", Resource: "providers",
	}},
	{Category: "function", GVR: schema.GroupVersionResource{
		Group: "pkg.crossplane.io", Version: "v1", Resource: "functions",
	}},
}

// CrossplaneResourcesHandler returns GET /api/v1/crossplane/resources, the
// aggregated list of Compositions, Providers, Functions, Managed Resources
// and Composite Resources the calling user is allowed to see.
func CrossplaneResourcesHandler(logger *slog.Logger, factory CrossplaneFactory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}

		user, ok := requireUser(w, r)
		if !ok {
			return
		}
		userName, groups := user.Kubernetes()

		dyn, err := factory.Dynamic(userName, groups)
		if err != nil {
			logger.ErrorContext(r.Context(), "build dynamic client", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "client_build_failed")
			return
		}
		disc, err := factory.Discovery(userName, groups)
		if err != nil {
			logger.ErrorContext(r.Context(), "build discovery client", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "client_build_failed")
			return
		}

		// Discover managed + composite GVRs via categories set on each CRD.
		managedGVRs, compositeGVRs := discoverCategoryGVRs(r.Context(), logger, disc)

		groupsOut := make([]CrossplaneGroup, 0, len(topLevelGVRs)+2)
		for _, top := range topLevelGVRs {
			groupsOut = append(
				groupsOut,
				listAsGroup(r.Context(), dyn, top.Category, []schema.GroupVersionResource{top.GVR}),
			)
		}
		groupsOut = append(groupsOut,
			listAsGroup(r.Context(), dyn, "managed", managedGVRs),
			listAsGroup(r.Context(), dyn, "composite", compositeGVRs),
		)

		writeJSON(w, logger, CrossplaneSummary{Groups: groupsOut})
	})
}

// listAsGroup lists every GVR in gvrs, merges the items and produces a
// single CrossplaneGroup. Per-GVR errors are swallowed except when *every*
// GVR for the category fails, in which case the first error surfaces as the
// group's Error field.
func listAsGroup(
	ctx context.Context,
	dyn dynamic.Interface,
	category string,
	gvrs []schema.GroupVersionResource,
) CrossplaneGroup {
	g := CrossplaneGroup{Category: category, Items: []CrossplaneResource{}}
	if len(gvrs) == 0 {
		return g
	}
	var firstErr error
	okCount := 0
	for _, gvr := range gvrs {
		list, err := dyn.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err != nil {
			// A missing CRD (404) or not-allowed (403) is not fatal for the
			// category; skip and note the error only if nothing else worked.
			if apierrors.IsNotFound(err) || apierrors.IsForbidden(err) || meta(err) {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		okCount++
		for i := range list.Items {
			g.Items = append(g.Items, projectResource(&list.Items[i], gvr.Resource))
		}
	}
	if okCount == 0 && firstErr != nil {
		g.Error = firstErr.Error()
	}
	sort.SliceStable(g.Items, func(i, j int) bool {
		if g.Items[i].Kind != g.Items[j].Kind {
			return g.Items[i].Kind < g.Items[j].Kind
		}
		return g.Items[i].Name < g.Items[j].Name
	})
	return g
}

// meta guards against the "no matches for kind" error the RESTMapper raises
// when a CRD is missing from the cluster.
func meta(err error) bool {
	return strings.Contains(err.Error(), "the server could not find the requested resource") ||
		strings.Contains(err.Error(), "no matches for kind")
}

func projectResource(u *unstructured.Unstructured, resource string) CrossplaneResource {
	ready, synced := extractConditions(u)
	return CrossplaneResource{
		APIVersion:        u.GetAPIVersion(),
		Kind:              u.GetKind(),
		Resource:          resource,
		Name:              u.GetName(),
		Namespace:         u.GetNamespace(),
		Ready:             ready,
		Synced:            synced,
		CreationTimestamp: u.GetCreationTimestamp().Time,
	}
}

// extractConditions walks status.conditions for the standard Crossplane
// Ready/Synced entries. Missing conditions collapse to "Unknown" so the UI
// always has a deterministic badge to render.
func extractConditions(u *unstructured.Unstructured) (ready, synced string) {
	ready, synced = "Unknown", "Unknown"
	conds, found, err := unstructured.NestedSlice(u.Object, "status", "conditions")
	if err != nil || !found {
		return ready, synced
	}
	for _, raw := range conds {
		c, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		t, _ := c["type"].(string)
		s, _ := c["status"].(string)
		switch t {
		case "Ready":
			ready = s
		case "Synced":
			synced = s
		}
	}
	return ready, synced
}

// discoverCategoryGVRs enumerates the cluster's API groups and returns the
// list of GVRs belonging to the "managed" and "composite" Crossplane
// categories. A single concurrent map guards the two slices so we only walk
// discovery once.
func discoverCategoryGVRs(
	ctx context.Context,
	logger *slog.Logger,
	disc discovery.DiscoveryInterface,
) (managed, composite []schema.GroupVersionResource) {
	_ = ctx // discovery client is pre-impersonated; no per-request ctx propagation here.
	apiGroups, apiResources, err := disc.ServerGroupsAndResources()
	if err != nil {
		// ServerGroupsAndResources returns partial results even on error.
		// Log and continue with what we got.
		logger.WarnContext(ctx, "discovery returned partial data", slog.String("error", err.Error()))
	}
	_ = apiGroups
	collect := func(category string, gvr schema.GroupVersionResource) {
		switch category {
		case "managed":
			managed = append(managed, gvr)
		case "composite":
			composite = append(composite, gvr)
		}
	}
	for _, group := range apiResources {
		gv, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			continue
		}
		for i := range group.APIResources {
			r := &group.APIResources[i]
			// Skip subresources (e.g. /status, /scale) — they carry a slash.
			if strings.Contains(r.Name, "/") {
				continue
			}
			for _, cat := range r.Categories {
				switch cat {
				case "managed":
					collect("managed", gv.WithResource(r.Name))
				case "composite":
					collect("composite", gv.WithResource(r.Name))
				}
			}
		}
	}
	return managed, composite
}
