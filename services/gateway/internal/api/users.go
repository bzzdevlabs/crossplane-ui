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
	"k8s.io/client-go/dynamic"
)

// userGVR / groupGVR are the two auth CRs the UI CRUDs.
var (
	userGVR = schema.GroupVersionResource{
		Group:    "auth.crossplane-ui.io",
		Version:  "v1alpha1",
		Resource: "users",
	}
	groupGVR = schema.GroupVersionResource{
		Group:    "auth.crossplane-ui.io",
		Version:  "v1alpha1",
		Resource: "groups",
	}
)

// UsersHandler wires GET / PUT / DELETE on /api/v1/auth/users.
func UsersHandler(logger *slog.Logger, factory CrossplaneFactory) http.Handler {
	return clusterScopedCRUD(logger, factory, userGVR, "User")
}

// GroupsHandler wires GET / PUT / DELETE on /api/v1/auth/groups.
func GroupsHandler(logger *slog.Logger, factory CrossplaneFactory) http.Handler {
	return clusterScopedCRUD(logger, factory, groupGVR, "Group")
}

// clusterScopedCRUD returns a handler that GET lists / GET single (?name=) /
// PUT server-side-applies / DELETE for a fixed cluster-scoped GVR.
func clusterScopedCRUD(
	logger *slog.Logger,
	factory CrossplaneFactory,
	gvr schema.GroupVersionResource,
	kind string,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		switch r.Method {
		case http.MethodGet:
			handleClusterGet(w, r, logger, dyn, gvr)
		case http.MethodPut:
			handleClusterApply(w, r, logger, dyn, gvr, kind)
		case http.MethodDelete:
			handleClusterDelete(w, r, logger, dyn, gvr)
		default:
			w.Header().Set("Allow", "GET, PUT, DELETE")
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		}
	})
}

func handleClusterGet(
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	dyn dynamic.Interface,
	gvr schema.GroupVersionResource,
) {
	single := strings.TrimSpace(r.URL.Query().Get("name"))
	if single != "" {
		obj, err := dyn.Resource(gvr).Get(r.Context(), single, metav1.GetOptions{})
		if err != nil {
			writeKubeError(w, logger, r, err, "get")
			return
		}
		writeJSON(w, logger, obj)
		return
	}
	list, err := dyn.Resource(gvr).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeKubeError(w, logger, r, err, "list")
		return
	}
	writeJSON(w, logger, list)
}

func handleClusterApply(
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	dyn dynamic.Interface,
	gvr schema.GroupVersionResource,
	kind string,
) {
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
	obj.SetAPIVersion(gvr.GroupVersion().String())
	obj.SetKind(kind)

	payload, err := json.Marshal(obj.Object)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "marshal_failed")
		return
	}
	opts := metav1.PatchOptions{FieldManager: FieldManager, Force: boolPtr(true)}
	applied, err := dyn.Resource(gvr).Patch(r.Context(), obj.GetName(), types.ApplyPatchType, payload, opts)
	if err != nil {
		writeKubeError(w, logger, r, err, "apply")
		return
	}
	writeJSON(w, logger, applied)
}

func handleClusterDelete(
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	dyn dynamic.Interface,
	gvr schema.GroupVersionResource,
) {
	single := strings.TrimSpace(r.URL.Query().Get("name"))
	if single == "" {
		writeError(w, http.StatusBadRequest, "missing 'name' query parameter")
		return
	}
	if err := dyn.Resource(gvr).Delete(r.Context(), single, metav1.DeleteOptions{}); err != nil {
		writeKubeError(w, logger, r, err, "delete")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UserPasswordRequest is the JSON body POSTed to /api/v1/auth/user-password.
// The auth controller watches the Secret and hashes the plaintext on its
// next reconcile, then blanks the key.
type UserPasswordRequest struct {
	Namespace string `json:"namespace"`
	// SecretName is the name of the Secret to upsert. The UI derives it from
	// User.spec.passwordSecretRef.name (or generates "user-<username>" when
	// creating).
	SecretName string `json:"secretName"`
	Password   string `json:"password"`
}

// UserPasswordHandler writes (upserts) a password Secret the auth controller
// will hash into User.status.passwordHash on the next reconcile.
func UserPasswordHandler(logger *slog.Logger, factory ClientFactory) http.Handler {
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
		var req UserPasswordRequest
		if err := json.NewDecoder(io.LimitReader(r.Body, MaxResourceBodySize)).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json: "+err.Error())
			return
		}
		if req.Namespace == "" || req.SecretName == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "namespace, secretName and password are required")
			return
		}

		kName, groups := user.Kubernetes()
		cs, err := factory.For(kName, groups)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "client_build_failed")
			return
		}

		desired := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.SecretName,
				Namespace: req.Namespace,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "crossplane-ui-gateway",
					"auth.crossplane-ui.io/kind":   "user",
				},
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{"password": []byte(req.Password)},
		}

		existing, gerr := cs.CoreV1().Secrets(req.Namespace).Get(r.Context(), req.SecretName, metav1.GetOptions{})
		if gerr == nil {
			existing.Data = desired.Data
			if existing.Labels == nil {
				existing.Labels = map[string]string{}
			}
			for k, v := range desired.Labels {
				existing.Labels[k] = v
			}
			// Clear the controller's consumed annotation so the next reconcile
			// re-hashes the fresh plaintext.
			if existing.Annotations != nil {
				delete(existing.Annotations, "auth.crossplane-ui.io/password-consumed")
			}
			updated, uerr := cs.CoreV1().Secrets(req.Namespace).Update(
				r.Context(), existing, metav1.UpdateOptions{FieldManager: FieldManager},
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
