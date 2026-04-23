// Package manager assembles the controller-runtime Manager used by the auth
// service.
//
// The manager hosts the User and Group reconcilers and owns the long-lived
// Kubernetes client shared with bootstrap logic before the leader-elected
// loops start.
package manager

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/config"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/controller"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/kube"
)

// Options collects what New needs besides the REST config.
type Options struct {
	RestConfig *rest.Config
	Config     *config.Config
	Logger     logr.Logger
}

// New builds and returns a ready-to-start controller-runtime Manager with the
// User and Group reconcilers registered.
func New(opts Options) (manager.Manager, error) {
	if opts.RestConfig == nil {
		return nil, errors.New("manager: nil rest.Config")
	}
	if opts.Config == nil {
		return nil, errors.New("manager: nil config.Config")
	}

	mgrOpts := manager.Options{
		Scheme:                  kube.Scheme,
		LeaderElection:          opts.Config.LeaderElection,
		LeaderElectionID:        "crossplane-ui-auth.leaderelection",
		LeaderElectionNamespace: opts.Config.Namespace,
		Metrics: metricsserver.Options{
			BindAddress: opts.Config.MetricsAddr,
		},
		HealthProbeBindAddress: opts.Config.HealthProbeAddr,
		Logger:                 opts.Logger,
		Cache: cache.Options{
			// Narrow the Secret informer so we don't stream unrelated objects
			// from the rest of the cluster. Users and Groups are cluster-scoped
			// and uncapped on purpose.
			ByObject: map[client.Object]cache.ByObject{
				&corev1.Secret{}: {
					Namespaces: map[string]cache.Config{
						opts.Config.Namespace: {},
					},
				},
			},
		},
	}

	mgr, err := ctrl.NewManager(opts.RestConfig, mgrOpts)
	if err != nil {
		return nil, fmt.Errorf("create manager: %w", err)
	}

	userR := &controller.UserReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Namespace: opts.Config.Namespace,
	}
	if err := userR.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("setup user reconciler: %w", err)
	}

	groupR := &controller.GroupReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	if err := groupR.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("setup group reconciler: %w", err)
	}

	return mgr, nil
}

// NewDirectClient builds an uncached client against the given config. It is
// used by one-shot startup work (bootstrap) that runs before the manager
// cache is populated.
func NewDirectClient(cfg *rest.Config) (client.Client, error) {
	return client.New(cfg, client.Options{Scheme: kube.Scheme})
}
