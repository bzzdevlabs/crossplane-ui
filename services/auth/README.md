# auth

The local-user and bootstrap service of **crossplane-ui**. The service:

- reconciles the `User` and `Group` custom resources defined by this project;
- bootstraps the administrator account the first time the chart is installed
  by reading a Kubernetes `Secret` and creating the corresponding `User`;
- projects every enabled `User` into a [Dex](https://dexidp.io/)
  `Password.dex.coreos.com/v1` object so Dex's password DB picks it up from
  its Kubernetes storage at request time — no Dex restart required;
- upserts the gateway's `OAuth2Client.dex.coreos.com/v1` object at startup
  from the shared `{release}-oidc` Secret, so the chart install materialises a
  ready-to-login OIDC client.

## Configuration

| Environment variable              | Default                              | Description                                                    |
| --------------------------------- | ------------------------------------ | -------------------------------------------------------------- |
| `HTTP_ADDR`                       | `:8081`                              | Listen address for the operational HTTP endpoints.             |
| `METRICS_ADDR`                    | `:8082`                              | Listen address for the controller-runtime metrics endpoint.    |
| `LOG_LEVEL`                       | `info`                               | One of `debug`, `info`, `warn`, `error`.                       |
| `LOG_FORMAT`                      | `json`                               | One of `json`, `text`.                                         |
| `KUBECONFIG`                      | _(empty)_                            | Empty means in-cluster auth.                                   |
| `POD_NAMESPACE`                   | _SA token_                           | Namespace where Secrets and ConfigMaps are read/written.       |
| `BOOTSTRAP_ADMIN_USERNAME`        | `admin`                              | Username of the bootstrap admin.                               |
| `BOOTSTRAP_ADMIN_PASSWORD_SECRET` | `crossplane-ui-bootstrap-admin`      | Secret that holds the bootstrap admin password (`password` key).|
| `OIDC_CLIENT_ID`                  | _(empty)_                            | Dex OAuth2 client id; skipped when empty.                      |
| `OIDC_CLIENT_NAME`                | `crossplane-ui`                      | Human-friendly name for the OAuth2Client CR.                   |
| `OIDC_CLIENT_SECRET_NAME`         | _(empty)_                            | Secret carrying the OAuth2 client secret.                      |
| `OIDC_CLIENT_SECRET_KEY`          | `clientSecret`                       | Key inside `OIDC_CLIENT_SECRET_NAME`.                          |
| `OIDC_REDIRECT_URIS`              | _(empty)_                            | Comma or whitespace separated list of allowed callback URLs.   |
| `LEADER_ELECTION`                 | `true`                               | Toggles controller-runtime leader election.                    |

## Local run

```bash
cd services/auth
go run ./cmd
```

## Layout

```
services/auth/
├── cmd/                 Entry point (main.go)
└── internal/
    ├── bootstrap/       First-run admin creation (Secret → User + bcrypt)
    ├── buildinfo/       Version metadata injected via ldflags
    ├── config/          Environment parsing
    ├── controller/      User and Group reconcilers (controller-runtime)
    ├── dex/             Projection of User CRs to Dex Password + OAuth2Client CRs
    ├── kube/            Runtime scheme + REST config loader
    ├── logging/         slog logger factory
    ├── manager/         Controller-runtime Manager setup
    ├── password/        bcrypt hash / verify helpers
    └── server/          HTTP handlers and tests
```

The `User` and `Group` Go types live in
[`pkg/apis/auth/v1alpha1`](../../pkg/apis/auth/v1alpha1). Their CRD
manifests are generated from the source types into the chart's
[`crds/`](../../deploy/helm/crossplane-ui/crds) directory — Helm 3
auto-applies them on install. Regenerate with `task generate`.

## Milestones in this service

- **M1 — Foundations**: skeleton, healthz/readyz/metrics, graceful shutdown.
- **M3 — User CR**: `User` / `Group` CRDs, controller-runtime reconciler,
  bootstrap admin, bcrypt helpers, Dex ConfigMap writer.
- **M4 — End-to-end login**: Password/OAuth2Client projection into Dex's
  Kubernetes storage, auto-generated OIDC client secret, chart-managed CRDs.
- **M8 — SSO connectors**: Dex connector templates for LDAP / SAML / OIDC / GitHub.
