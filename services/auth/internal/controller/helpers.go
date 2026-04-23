package controller

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	authv1alpha1 "github.com/bzzdevlabs/crossplane-ui/pkg/apis/auth/v1alpha1"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/dex"
)

// setCondition writes a Condition of type Ready to conds. It creates the
// slice if nil and updates the LastTransitionTime through meta.SetStatusCondition.
func setCondition(conds *[]metav1.Condition, status metav1.ConditionStatus, reason, message string, generation int64) {
	if conds == nil {
		return
	}
	meta.SetStatusCondition(conds, metav1.Condition{
		Type:               authv1alpha1.ConditionReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: generation,
	})
}

// managedByPredicate accepts only events for objects labelled as managed by
// this controller. Used to filter Dex Password events so external CRs do not
// drive our reconcile loop.
type managedByPredicate struct{ predicate.Funcs }

func (managedByPredicate) match(o client.Object) bool {
	if o == nil {
		return false
	}
	return o.GetLabels()[dex.ManagedByLabel] == dex.ManagedByValue
}

func (p managedByPredicate) Create(e event.CreateEvent) bool { return p.match(e.Object) }
func (p managedByPredicate) Update(e event.UpdateEvent) bool {
	return p.match(e.ObjectNew) || p.match(e.ObjectOld)
}
func (p managedByPredicate) Delete(e event.DeleteEvent) bool   { return p.match(e.Object) }
func (p managedByPredicate) Generic(e event.GenericEvent) bool { return p.match(e.Object) }
