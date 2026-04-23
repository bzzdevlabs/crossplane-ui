package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// FieldManager is the identifier stamped on server-side-apply patches so the
// API server attributes changes to this application.
const FieldManager = "crossplane-ui-gateway"

// MaxResourceBodySize caps JSON bodies accepted by PUT handlers. Kubernetes
// itself rejects objects beyond a similar threshold; enforcing it here keeps
// pathological payloads from pinning the gateway.
const MaxResourceBodySize = 1 << 20 // 1 MiB

// resourceTarget carries the GVR + name + namespace parsed from query
// parameters.
type resourceTarget struct {
	GVR       schema.GroupVersionResource
	Namespace string
	Name      string
}

// parseResourceTarget reads the mandatory group/version/resource/name query
// parameters. Namespace is optional (cluster-scoped resources).
func parseResourceTarget(r *http.Request, requireName bool) (resourceTarget, error) {
	q := r.URL.Query()
	t := resourceTarget{
		GVR: schema.GroupVersionResource{
			Group:    q.Get("group"),
			Version:  q.Get("version"),
			Resource: q.Get("resource"),
		},
		Namespace: q.Get("namespace"),
		Name:      q.Get("name"),
	}
	if t.GVR.Version == "" {
		return t, errors.New("missing 'version' query parameter")
	}
	if t.GVR.Resource == "" {
		return t, errors.New("missing 'resource' query parameter")
	}
	if requireName && t.Name == "" {
		return t, errors.New("missing 'name' query parameter")
	}
	return t, nil
}

// ResourceHandler mounts the generic `/api/v1/crossplane/resource` endpoint
// that drives the list/detail/apply/delete loop from the UI. It relies on the
// impersonated dynamic client from the CrossplaneFactory so every call is
// evaluated against the caller's RBAC.
func ResourceHandler(logger *slog.Logger, factory CrossplaneFactory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleResourceGet(w, r, logger, factory)
		case http.MethodPut:
			handleResourceApply(w, r, logger, factory)
		case http.MethodDelete:
			handleResourceDelete(w, r, logger, factory)
		default:
			w.Header().Set("Allow", "GET, PUT, DELETE")
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		}
	})
}

func handleResourceGet(w http.ResponseWriter, r *http.Request, logger *slog.Logger, factory CrossplaneFactory) {
	dyn, target, ok := prepareRequest(w, r, logger, factory, true)
	if !ok {
		return
	}
	obj, err := dyn.Resource(target.GVR).Namespace(target.Namespace).Get(r.Context(), target.Name, metav1.GetOptions{})
	if err != nil {
		writeKubeError(w, logger, r, err, "get")
		return
	}
	writeJSON(w, logger, obj)
}

func handleResourceApply(w http.ResponseWriter, r *http.Request, logger *slog.Logger, factory CrossplaneFactory) {
	dyn, target, ok := prepareRequest(w, r, logger, factory, false)
	if !ok {
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, MaxResourceBodySize+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "read_body_failed")
		return
	}
	if len(body) > MaxResourceBodySize {
		writeError(w, http.StatusRequestEntityTooLarge, "body_too_large")
		return
	}
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal(body, &obj.Object); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json: "+err.Error())
		return
	}
	if obj.GetName() == "" {
		writeError(w, http.StatusBadRequest, "metadata.name is required")
		return
	}
	gv := schema.GroupVersion{Group: target.GVR.Group, Version: target.GVR.Version}
	obj.SetAPIVersion(gv.String())
	if target.Namespace != "" {
		obj.SetNamespace(target.Namespace)
	}
	name := target.Name
	if name == "" {
		name = obj.GetName()
	}

	dryRun := r.URL.Query().Get("dryRun") == "All"
	opts := metav1.PatchOptions{
		FieldManager: FieldManager,
		Force:        boolPtr(true),
	}
	if dryRun {
		opts.DryRun = []string{metav1.DryRunAll}
	}

	payload, err := json.Marshal(obj.Object)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "marshal_failed")
		return
	}
	applied, err := dyn.Resource(target.GVR).Namespace(target.Namespace).
		Patch(r.Context(), name, types.ApplyPatchType, payload, opts)
	if err != nil {
		writeKubeError(w, logger, r, err, "apply")
		return
	}
	writeJSON(w, logger, applied)
}

func handleResourceDelete(w http.ResponseWriter, r *http.Request, logger *slog.Logger, factory CrossplaneFactory) {
	dyn, target, ok := prepareRequest(w, r, logger, factory, true)
	if !ok {
		return
	}
	if err := dyn.Resource(target.GVR).Namespace(target.Namespace).
		Delete(r.Context(), target.Name, metav1.DeleteOptions{}); err != nil {
		writeKubeError(w, logger, r, err, "delete")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// prepareRequest authenticates, validates query params and builds the
// impersonated dynamic client. It writes the error response itself on
// failure; callers bail out when ok is false.
func prepareRequest(
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	factory CrossplaneFactory,
	requireName bool,
) (dynamic.Interface, resourceTarget, bool) {
	user, ok := requireUser(w, r)
	if !ok {
		return nil, resourceTarget{}, false
	}
	target, err := parseResourceTarget(r, requireName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return nil, resourceTarget{}, false
	}
	name, groups := user.Kubernetes()
	dyn, err := factory.Dynamic(name, groups)
	if err != nil {
		logger.ErrorContext(r.Context(), "build dynamic client", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "client_build_failed")
		return nil, resourceTarget{}, false
	}
	return dyn, target, true
}

func boolPtr(b bool) *bool { return &b }

func writeKubeError(w http.ResponseWriter, logger *slog.Logger, r *http.Request, err error, op string) {
	switch {
	case apierrors.IsNotFound(err):
		writeError(w, http.StatusNotFound, "not_found")
	case apierrors.IsForbidden(err):
		writeError(w, http.StatusForbidden, "forbidden")
	case apierrors.IsUnauthorized(err):
		writeError(w, http.StatusUnauthorized, "unauthorized")
	case apierrors.IsConflict(err):
		writeError(w, http.StatusConflict, "conflict")
	case apierrors.IsInvalid(err):
		writeError(w, http.StatusBadRequest, "invalid: "+err.Error())
	default:
		logger.WarnContext(r.Context(), "kube verb failed",
			slog.String("op", op), slog.String("error", err.Error()))
		writeError(w, http.StatusBadGateway, "kube_"+op+"_failed")
	}
}
