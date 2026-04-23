package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// connectorGVR is the Connector CRD the auth service controls.
var connectorGVR = schema.GroupVersionResource{
	Group:    "auth.crossplane-ui.io",
	Version:  "v1alpha1",
	Resource: "connectors",
}

// ConnectorsHandler wires GET / PUT / DELETE on /api/v1/auth/connectors.
// Cluster-scoped; the GVR is fixed so the UI never passes group/version/resource.
func ConnectorsHandler(logger *slog.Logger, factory CrossplaneFactory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleConnectorsGet(w, r, logger, factory)
		case http.MethodPut:
			handleConnectorApply(w, r, logger, factory)
		case http.MethodDelete:
			handleConnectorDelete(w, r, logger, factory)
		default:
			w.Header().Set("Allow", "GET, PUT, DELETE")
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		}
	})
}

func handleConnectorsGet(w http.ResponseWriter, r *http.Request, logger *slog.Logger, factory CrossplaneFactory) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	name, groups := user.Kubernetes()
	dyn, err := factory.Dynamic(name, groups)
	if err != nil {
		logger.ErrorContext(r.Context(), "build dynamic client", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "client_build_failed")
		return
	}

	single := strings.TrimSpace(r.URL.Query().Get("name"))
	if single != "" {
		obj, err := dyn.Resource(connectorGVR).Get(r.Context(), single, metav1.GetOptions{})
		if err != nil {
			writeKubeError(w, logger, r, err, "get")
			return
		}
		writeJSON(w, logger, obj)
		return
	}

	list, err := dyn.Resource(connectorGVR).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeKubeError(w, logger, r, err, "list")
		return
	}
	writeJSON(w, logger, list)
}

func handleConnectorApply(w http.ResponseWriter, r *http.Request, logger *slog.Logger, factory CrossplaneFactory) {
	user, ok := requireUser(w, r)
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
	obj.SetAPIVersion(connectorGVR.GroupVersion().String())
	obj.SetKind("Connector")

	kName, groups := user.Kubernetes()
	dyn, err := factory.Dynamic(kName, groups)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "client_build_failed")
		return
	}

	payload, err := json.Marshal(obj.Object)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "marshal_failed")
		return
	}
	opts := metav1.PatchOptions{FieldManager: FieldManager, Force: boolPtr(true)}
	applied, err := dyn.Resource(connectorGVR).Patch(r.Context(), obj.GetName(), types.ApplyPatchType, payload, opts)
	if err != nil {
		writeKubeError(w, logger, r, err, "apply")
		return
	}
	writeJSON(w, logger, applied)
}

func handleConnectorDelete(w http.ResponseWriter, r *http.Request, logger *slog.Logger, factory CrossplaneFactory) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		writeError(w, http.StatusBadRequest, "missing 'name' query parameter")
		return
	}
	kName, groups := user.Kubernetes()
	dyn, err := factory.Dynamic(kName, groups)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "client_build_failed")
		return
	}
	if err := dyn.Resource(connectorGVR).Delete(r.Context(), name, metav1.DeleteOptions{}); err != nil {
		writeKubeError(w, logger, r, err, "delete")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ConnectorSecretRequest is the JSON body POSTed to /api/v1/auth/connector-secrets.
type ConnectorSecretRequest struct {
	// Namespace must be the auth controller's namespace (where it reads
	// connector Secrets from). The handler does not default it; the UI
	// surfaces the gateway's /api/v1/config.authNamespace for this.
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	Data      map[string]string `json:"data"`
}

// ConnectorSecretsHandler writes (upserts) a Secret used by a Connector's
// secretRefs. The caller's impersonated identity must carry secrets
// create/update RBAC in the target namespace.
func ConnectorSecretsHandler(logger *slog.Logger, factory ClientFactory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.Header().Set("Allow", http.MethodPut)
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		user, ok := requireUser(w, r)
		if !ok {
			return
		}
		var req ConnectorSecretRequest
		if err := json.NewDecoder(io.LimitReader(r.Body, MaxResourceBodySize)).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json: "+err.Error())
			return
		}
		if req.Namespace == "" || req.Name == "" || len(req.Data) == 0 {
			writeError(w, http.StatusBadRequest, "namespace, name and data are required")
			return
		}

		kName, groups := user.Kubernetes()
		cs, err := factory.For(kName, groups)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "client_build_failed")
			return
		}

		data := make(map[string][]byte, len(req.Data))
		for k, v := range req.Data {
			data[k] = []byte(v)
		}

		desired := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "crossplane-ui-gateway",
					"auth.crossplane-ui.io/kind":   "connector",
				},
			},
			Type: corev1.SecretTypeOpaque,
			Data: data,
		}

		sec, err := cs.CoreV1().Secrets(req.Namespace).Get(r.Context(), req.Name, metav1.GetOptions{})
		if err == nil {
			sec.Data = desired.Data
			if sec.Labels == nil {
				sec.Labels = map[string]string{}
			}
			for k, v := range desired.Labels {
				sec.Labels[k] = v
			}
			updated, uerr := cs.CoreV1().Secrets(req.Namespace).Update(
				r.Context(), sec, metav1.UpdateOptions{FieldManager: FieldManager},
			)
			if uerr != nil {
				writeKubeError(w, logger, r, uerr, "update")
				return
			}
			writeJSON(w, logger, projectSecret(updated))
			return
		}
		created, cerr := cs.CoreV1().Secrets(req.Namespace).Create(
			r.Context(), desired, metav1.CreateOptions{FieldManager: FieldManager},
		)
		if cerr != nil {
			writeKubeError(w, logger, r, cerr, "create")
			return
		}
		writeJSON(w, logger, projectSecret(created))
	})
}

// projectSecret strips data before returning so a write response never echoes
// the plaintext back over the wire.
func projectSecret(s *corev1.Secret) map[string]any {
	keys := make([]string, 0, len(s.Data))
	for k := range s.Data {
		keys = append(keys, k)
	}
	return map[string]any{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]any{
			"name":      s.Name,
			"namespace": s.Namespace,
			"labels":    s.Labels,
		},
		"keys": keys,
	}
}
