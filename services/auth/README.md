# auth

The local-user and bootstrap service of **crossplane-ui**. The service:

- reconciles the `User` and `Group` custom resources defined by this project;
- bootstraps the administrator account the first time the chart is installed
  by reading a Kubernetes `Secret` and creating the corresponding `User`;
- keeps the [Dex](https://dexidp.io/) static password DB in sync with the
  `User` CRs so that Dex can authenticate local users.

## Configuration

| Environment variable              | Default                              | Description                                                    |
| --------------------------------- | ------------------------------------ | -------------------------------------------------------------- |
| `HTTP_ADDR`                       | `:8081`                              | Listen address.                                                |
| `LOG_LEVEL`                       | `info`                               | One of `debug`, `info`, `warn`, `error`.                       |
| `LOG_FORMAT`                      | `json`                               | One of `json`, `text`.                                         |
| `KUBECONFIG`                      | _(empty)_                            | Empty means in-cluster auth.                                   |
| `BOOTSTRAP_ADMIN_USERNAME`        | `admin`                              | Username of the bootstrap admin.                               |
| `BOOTSTRAP_ADMIN_PASSWORD_SECRET` | `crossplane-ui-bootstrap-admin`      | Secret that holds the bootstrap admin password (`password` key).|
| `DEX_CONFIGMAP_NAME`              | `crossplane-ui-dex-config`           | ConfigMap rewritten by this service to update Dex users.       |

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
    ├── buildinfo/       Version metadata injected via ldflags
    ├── config/          Environment parsing
    ├── logging/         slog logger factory
    └── server/          HTTP handlers and tests
```

## Milestones in this service

- **M1 — Foundations**: skeleton, healthz/readyz/metrics, graceful shutdown.
- **M3 — User CR**: `User` / `Group` CRDs, controller-runtime reconciler,
  bootstrap admin, bcrypt helpers, Dex ConfigMap writer.
- **M8 — SSO connectors**: Dex connector templates for LDAP / SAML / OIDC / GitHub.
