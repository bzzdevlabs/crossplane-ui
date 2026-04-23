package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	authv1alpha1 "github.com/bzzdevlabs/crossplane-ui/pkg/apis/auth/v1alpha1"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/dex"
)

const (
	reasonConnectorProjected = "ProjectedToDex"
	reasonConnectorDisabled  = "Disabled"
	reasonConnectorError     = "ProjectionError"
)

// ConnectorReconciler projects Connector CRs into Dex Connector storage.
type ConnectorReconciler struct {
	Client    client.Client
	Scheme    *runtime.Scheme
	Namespace string
}

// SetupWithManager wires the reconciler and re-enqueues on Dex Connector
// events so outside edits get overwritten.
func (r *ConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	connList := &unstructured.Unstructured{}
	connList.SetGroupVersionKind(dex.ConnectorGVK)

	return ctrl.NewControllerManagedBy(mgr).
		Named("connector").
		For(&authv1alpha1.Connector{}).
		WatchesRawSource(
			source.Kind(
				mgr.GetCache(),
				client.Object(connList),
				handler.EnqueueRequestsFromMapFunc(r.mapConnectorToCRs),
				managedByPredicate{},
			),
		).
		Complete(r)
}

// Reconcile implements reconcile.Reconciler.
func (r *ConnectorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx).WithValues("connector", req.Name)

	var c authv1alpha1.Connector
	if err := r.Client.Get(ctx, req.NamespacedName, &c); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, r.syncDex(ctx, logger)
		}
		return reconcile.Result{}, fmt.Errorf("get connector: %w", err)
	}

	patch := client.MergeFrom(c.DeepCopy())
	c.Status.ObservedGeneration = c.Generation

	if err := r.syncDex(ctx, logger); err != nil {
		setCondition(&c.Status.Conditions, metav1.ConditionFalse, reasonConnectorError, err.Error(), c.Generation)
		_ = r.Client.Status().Patch(ctx, &c, patch)
		return reconcile.Result{}, err
	}

	switch {
	case c.Spec.Disabled:
		setCondition(&c.Status.Conditions, metav1.ConditionFalse, reasonConnectorDisabled,
			"connector disabled; pruned from Dex storage", c.Generation)
	default:
		setCondition(&c.Status.Conditions, metav1.ConditionTrue, reasonConnectorProjected,
			"connector projected into Dex storage", c.Generation)
	}

	if err := r.Client.Status().Patch(ctx, &c, patch); err != nil {
		return reconcile.Result{}, fmt.Errorf("patch connector status: %w", err)
	}
	return reconcile.Result{}, nil
}

func (r *ConnectorReconciler) syncDex(ctx context.Context, logger logr.Logger) error {
	var list authv1alpha1.ConnectorList
	if err := r.Client.List(ctx, &list); err != nil {
		return fmt.Errorf("list connectors: %w", err)
	}
	changed, err := dex.SyncConnectors(ctx, r.Client, dex.KubeSecretResolver(r.Client), r.Namespace, list.Items)
	if err != nil {
		logger.Error(err, "dex connector sync failed", "namespace", r.Namespace)
		return err
	}
	if changed {
		logger.Info("dex connector objects reconciled",
			"namespace", r.Namespace,
			"connectors", len(list.Items))
	}
	return nil
}

func (r *ConnectorReconciler) mapConnectorToCRs(ctx context.Context, _ client.Object) []reconcile.Request {
	var list authv1alpha1.ConnectorList
	if err := r.Client.List(ctx, &list); err != nil {
		return nil
	}
	out := make([]reconcile.Request, 0, len(list.Items))
	for i := range list.Items {
		out = append(out, reconcile.Request{NamespacedName: types.NamespacedName{Name: list.Items[i].Name}})
	}
	if len(out) == 0 {
		out = append(out, reconcile.Request{NamespacedName: types.NamespacedName{Name: "__rebuild__"}})
	}
	return out
}
