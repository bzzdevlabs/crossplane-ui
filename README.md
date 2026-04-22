# crossplane-ui

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go version](https://img.shields.io/badge/go-1.26-00ADD8.svg)](https://go.dev/)
[![Vue 3](https://img.shields.io/badge/Vue-3.5-42b883.svg)](https://vuejs.org/)
[![Crossplane](https://img.shields.io/badge/Crossplane-v2-8B5CF6.svg)](https://crossplane.io/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.33+-326CE5.svg)](https://kubernetes.io/)
[![Status: alpha](https://img.shields.io/badge/status-alpha-orange.svg)](#project-status)

An admin UI for [Crossplane](https://crossplane.io) v2 — visualize, create, edit
and delete Crossplane resources (Compositions, Composite Resources, Managed
Resources, Providers, Functions, …) either through a YAML editor or through
Rancher-inspired configuration views.

> **Status — alpha.** This project is in active early development. APIs,
> CRDs and chart values are subject to breaking changes until v1.0.0.

## Features

- **Rancher-inspired UI** — card/tile dashboard of every deployed composition
  and resource with aggregated sync / ready status.
- **Dual editing mode** — either a rich YAML editor with schema validation, or
  Rancher-style configuration forms for each resource kind.
- **Authentication out of the box** — local users (bootstrap admin at install
  time) plus OIDC / OAuth / SAML / LDAP via [Dex](https://dexidp.io/).
- **Kubernetes-native RBAC** — every user request is forwarded to the Kubernetes
  API via [user impersonation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation);
  permissions are expressed as regular `(Cluster)RoleBinding`s.
- **Cloud-native deployment** — delivered as a Helm chart, ships as lightweight
  distroless container images.
- **Observability by default** — Prometheus `/metrics` on every service and
  structured JSON logs via [`log/slog`](https://pkg.go.dev/log/slog).
- **Bilingual UI (FR / EN)** — first-class internationalisation via
  [`vue-i18n`](https://vue-i18n.intlify.dev/).

## Architecture

```
                    ┌─────────────────────────────────────────┐
                    │   Kubernetes cluster (w/ Crossplane v2) │
                    └─────────────────────────────────────────┘
                                       │
  ┌────────────────────────────────────┼─────────────────────────────────┐
  │                                    ▼                                 │
  │  ┌─────────────┐    ┌─────────────────────────────┐   ┌───────────┐  │
  │  │   Browser   │───▶│  gateway  (Go)              │◀─▶│  kube-api │  │
  │  │  Vue 3 UI   │    │  - serves embedded UI        │   └───────────┘  │
  │  └─────────────┘    │  - OIDC verify (via Dex)     │         ▲        │
  │         ▲           │  - K8s user impersonation    │         │        │
  │         │           │  - Crossplane domain API     │         │ impersonate
  │         │           └────────┬─────────────┬───────┘         │        │
  │         │                    │             │                 │        │
  │         │            ┌───────▼────┐   ┌────▼──────┐          │        │
  │         │            │   dex      │   │  auth (Go)│──────────┘        │
  │         └────────────│ OIDC/LDAP/ │◀──│ users CRD │                   │
  │             login    │ SAML/OAuth │   │ bootstrap │                   │
  │                      └────────────┘   └───────────┘                   │
  └───────────────────────────────────────────────────────────────────────┘
```

See [docs/architecture.md](docs/architecture.md) for the detailed description
and [docs/adr/](docs/adr/) for the architecture decision records.

## Quick start (local dev)

Prerequisites:

- [Go](https://go.dev/) 1.26+
- [Node.js](https://nodejs.org/) 22+ and [pnpm](https://pnpm.io/) 10+
- [Task](https://taskfile.dev/) 3+ (task runner used by this repo)
- [Helm](https://helm.sh/) 4+
- A Kubernetes cluster with Crossplane v2 installed (kind / k3d / minikube work)

Bootstrap a local dev environment:

```bash
task bootstrap   # installs Go tools and pnpm dependencies
task lint        # runs all linters
task test        # runs all tests
task dev         # starts gateway, auth and the Vue dev server with hot reload
```

Deploy the chart to your cluster:

```bash
helm dependency update deploy/helm/crossplane-ui
helm upgrade --install crossplane-ui deploy/helm/crossplane-ui \
    --namespace crossplane-ui --create-namespace \
    --values deploy/helm/crossplane-ui/values.yaml
```

See [docs/deployment.md](docs/deployment.md) for a full deployment guide, and
[docs/development.md](docs/development.md) for the developer workflow.

## Repository layout

```
.
├── services/
│   ├── gateway/      Go — HTTP gateway, OIDC middleware, K8s impersonation
│   └── auth/         Go — local user controller, admin bootstrap, Dex sync
├── web/ui/           Vue 3.5 + Vite 8 + Pinia + vue-i18n + @rancher/components
├── deploy/helm/      Helm chart (depends on the upstream Dex chart)
├── pkg/              Shared Go libraries
├── docs/             Architecture, deployment, ADRs, …
└── test/             Integration and end-to-end suites
```

## Project status

This project is currently at milestone **M1 — Foundations**. The roadmap is
tracked in [docs/roadmap.md](docs/roadmap.md).

## Contributing

Contributions are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md), and
please read our [Code of Conduct](CODE_OF_CONDUCT.md) and
[Security Policy](SECURITY.md).

## License

Licensed under the [Apache License, Version 2.0](LICENSE).
