# Architecture

## Overview

**crossplane-ui** is a three-component application deployed as a Helm release
into a Kubernetes cluster that already runs [Crossplane v2](https://crossplane.io).

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

## Components

### 1. `gateway` (Go)

The HTTP entrypoint hit by the browser. It:

- serves the embedded Vue bundle on `/`;
- exposes a REST API under `/api/v1/` that wraps the Kubernetes API with
  Crossplane-aware endpoints (aggregated sync/ready status, composition
  hierarchy walks, YAML validation);
- enforces authentication by verifying OIDC ID tokens minted by Dex;
- propagates the authenticated identity to the Kubernetes API via the
  [user impersonation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation)
  headers (`Impersonate-User`, `Impersonate-Group`), so authorization is
  handled by plain Kubernetes `(Cluster)RoleBindings`.

It is **stateless** and deployed with 2 replicas behind a ClusterIP Service.

### 2. `auth` (Go)

A tiny Kubernetes controller + bootstrap service. It:

- reconciles the `User` and `Group` CRDs defined by this project;
- bootstraps the administrator at first install from a `Secret`;
- keeps Dex's static password DB in sync with the `User` CRs.

Local users are stored as `User` CRs with their bcrypt hash in a companion
`Secret` that only the `auth` ServiceAccount can read.

It is a **single-leader** controller; high availability via leader election is
on the M3 roadmap.

### 3. `dex`

Upstream [Dex IdP](https://dexidp.io/) chart, pulled as a dependency. Dex is
the sole identity provider surface exposed to the browser and to the gateway.
It federates any combination of:

- local users (passwordDB — fed by our `auth` service);
- LDAP / Active Directory;
- SAML 2.0;
- OIDC upstreams (Google, Azure AD, Keycloak, …);
- OAuth (GitHub, GitLab, …).

The gateway is a confidential OIDC client of Dex.

### 4. `web/ui`

A Vue 3 + Vite SPA. Built as a static bundle embedded into the `gateway`
binary at release time. Dev mode proxies `/api` to a running gateway.

## Request flow (post-login)

1. Browser sends `GET /api/v1/compositions` to the gateway with the ID token in
   a cookie (or `Authorization: Bearer ...`).
2. Gateway verifies the token against Dex's public keys (cached; JWK rotation
   honoured).
3. Gateway builds a Kubernetes request, adds
   `Impersonate-User: <sub>` and `Impersonate-Group: ...` headers, and forwards
   it to the Kubernetes API using the gateway's own `ServiceAccount`.
4. Kubernetes evaluates RBAC **as the impersonated user**; the request fails
   with `403` if the user does not have the necessary RoleBindings.
5. Gateway post-processes the response (status aggregation, sparse fields) and
   returns JSON to the browser.

## Trust boundaries

| Boundary                            | Protection                                                                 |
| ----------------------------------- | -------------------------------------------------------------------------- |
| Browser ↔ gateway                   | HTTPS (cert-manager recommended); HttpOnly cookie for the refresh token.   |
| Gateway ↔ Kubernetes API            | In-cluster mTLS (standard K8s).                                            |
| Gateway ↔ Dex (OIDC client)         | Client secret stored in a Kubernetes Secret.                               |
| Auth ↔ Kubernetes API               | `ServiceAccount` restricted via RBAC to User/Group CRs and the Dex CMap.   |
| Auth ↔ bootstrap Secret             | `get`/`update` only; `helm.sh/resource-policy: keep` preserves the Secret. |

## Observability

- Every service exposes `/metrics` (Prometheus text) and `/healthz`, `/readyz`.
- Structured logs via `log/slog` with a JSON handler.
- An optional `ServiceMonitor` is rendered by the chart when the user sets
  `serviceMonitor.enabled=true` (needs prometheus-operator).
- Distributed tracing via OTEL is on the roadmap (see
  [docs/adr/0002-observability.md](adr/0002-observability.md), planned).

## What this project explicitly is _not_

- **Not a Rancher replacement.** The UI reuses `@rancher/components` for
  look & feel but does not attempt to cover Rancher's scope.
- **Not multi-cluster (yet).** One deployment manages one Crossplane control
  plane. Multi-cluster is an accepted future extension.
- **Not a backup/DR tool** for Crossplane resources; use
  [Velero](https://velero.io/) or a GitOps workflow for that.
