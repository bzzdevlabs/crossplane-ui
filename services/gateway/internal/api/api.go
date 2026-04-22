// Package api exposes the HTTP handlers that back the gateway's REST API.
//
// Handlers in this package never talk to the Kubernetes API directly; they
// go through a ClientFactory that stamps the authenticated user's identity
// on every call via Kubernetes impersonation. This keeps authorization in
// the hands of the Kubernetes API server (plain RBAC).
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"k8s.io/client-go/kubernetes"

	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/oidc"
)

// ClientFactory is the narrow contract the api package consumes from
// internal/kube. Returning kubernetes.Interface (rather than the concrete
// *kubernetes.Clientset) lets tests plug in a `client-go/kubernetes/fake`
// clientset without any adapter.
type ClientFactory interface {
	For(user string, groups []string) (kubernetes.Interface, error)
}

// writeJSON marshals v as an application/json response with the given status
// code. Marshaling failures are logged and downgraded to 500.
func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	buf, err := json.Marshal(v)
	if err != nil {
		logger.Error("json marshal failed", slog.String("error", err.Error()))
		http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	_, _ = w.Write(buf)
}

// writeError emits a uniform JSON error body.
func writeError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}

// requireUser returns the authenticated user or writes a 401 and returns ok=false.
func requireUser(w http.ResponseWriter, r *http.Request) (oidc.User, bool) {
	u, ok := oidc.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return oidc.User{}, false
	}
	return u, true
}
