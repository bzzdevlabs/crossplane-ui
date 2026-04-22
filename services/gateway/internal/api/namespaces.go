package api

import (
	"log/slog"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Namespace is the gateway's external projection of a Kubernetes Namespace.
// We deliberately emit only the fields the UI consumes — enough to render a
// tile and navigate, not a full copy of the Kubernetes object.
type Namespace struct {
	Name              string            `json:"name"`
	Phase             string            `json:"phase"`
	Labels            map[string]string `json:"labels,omitempty"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
}

// NamespacesList is the envelope returned by GET /api/v1/namespaces.
type NamespacesList struct {
	Items []Namespace `json:"items"`
}

// NamespacesHandler returns GET /api/v1/namespaces, listing the cluster's
// namespaces as seen by the caller (impersonated).
func NamespacesHandler(logger *slog.Logger, factory ClientFactory) http.Handler {
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

		client, err := factory.For(userName, groups)
		if err != nil {
			logger.ErrorContext(r.Context(), "build impersonating client",
				slog.String("impersonate_user", userName),
				slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "client_build_failed")
			return
		}

		list, err := client.CoreV1().Namespaces().List(r.Context(), metav1.ListOptions{})
		if err != nil {
			logger.WarnContext(r.Context(), "kube list namespaces failed",
				slog.String("impersonate_user", userName),
				slog.String("error", err.Error()))
			writeError(w, http.StatusBadGateway, "kube_list_failed")
			return
		}

		out := make([]Namespace, 0, len(list.Items))
		for _, ns := range list.Items {
			out = append(out, Namespace{
				Name:              ns.Name,
				Phase:             string(ns.Status.Phase),
				Labels:            ns.Labels,
				CreationTimestamp: ns.CreationTimestamp.Time,
			})
		}

		writeJSON(w, logger, http.StatusOK, NamespacesList{Items: out})
	})
}
