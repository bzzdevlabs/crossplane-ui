# gateway

The HTTP entrypoint of **crossplane-ui**. The gateway:

- serves the embedded Vue UI (as static assets in later milestones);
- verifies OIDC ID tokens issued by [Dex](https://dexidp.io/);
- forwards API calls to the Kubernetes API with
  [user impersonation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation);
- exposes operational endpoints (`/healthz`, `/readyz`, `/metrics`).

## Middleware stack

Every request passes through, in order:

1. **Recover** — turns panics into 500s and logs the stack.
2. **RequestID** — attaches `X-Request-ID` to the request context and echoes
   it in the response.
3. **AccessLog** — emits one structured JSON record per request.
4. **Metrics** — records Prometheus counters and latency histograms.
5. **CORS** — allows the configured origins (see `CORS_ALLOWED_ORIGINS`).
6. **Auth** — OIDC verifier (or dev pass-through when disabled).

## Routes

| Method | Path                    | Auth    | Description                                     |
| ------ | ----------------------- | ------- | ----------------------------------------------- |
| GET    | `/healthz`              | public  | Liveness; always 200 when the process runs.     |
| GET    | `/readyz`               | public  | Readiness; 503 during startup.                  |
| GET    | `/metrics`              | public  | Prometheus text exposition.                     |
| GET    | `/api/v1/namespaces`    | yes     | Lists namespaces as seen by the caller.         |

## Configuration

| Environment variable       | Default                                                       | Description                                        |
| -------------------------- | ------------------------------------------------------------- | -------------------------------------------------- |
| `HTTP_ADDR`                | `:8080`                                                       | Listen address.                                    |
| `HTTP_READ_HEADER_TIMEOUT` | `10s`                                                         | `http.Server.ReadHeaderTimeout`.                   |
| `LOG_LEVEL`                | `info`                                                        | One of `debug`, `info`, `warn`, `error`.           |
| `LOG_FORMAT`               | `json`                                                        | One of `json`, `text`.                             |
| `KUBECONFIG`               | _(empty)_                                                     | Path to a kubeconfig; empty means in-cluster auth. |
| `OIDC_ISSUER_URL`          | _(empty)_                                                     | Expected `iss`. Empty → dev pass-through (no auth).|
| `OIDC_DISCOVERY_URL`       | `OIDC_ISSUER_URL`                                             | Override where discovery is fetched from.          |
| `OIDC_CLIENT_ID`           | _(empty)_                                                     | OIDC client id; required when OIDC is enabled.     |
| `OIDC_SKIP_ISSUER_CHECK`   | `false`                                                       | Dev-only: skip `iss` validation.                   |
| `AUTH_SERVICE_URL`         | `http://auth.crossplane-ui.svc.cluster.local:8081`            | Base URL of the auth service.                      |
| `CORS_ALLOWED_ORIGINS`     | `http://localhost:5173`                                       | Comma-separated list of accepted `Origin` values.  |

### Dev pass-through mode

When `OIDC_ISSUER_URL` is empty the gateway injects a synthetic admin user
(`dev-admin`, groups `system:masters`) on every request. This is convenient
for `curl` during development; it **MUST NOT** run in production. The process
emits a prominent warning log at startup to make the mode obvious.

## Local run

```bash
cd services/gateway
go run ./cmd
# then:
curl http://localhost:8080/healthz
curl http://localhost:8080/api/v1/namespaces
```

## Metrics

All metrics live under the `crossplane_ui_gateway_` prefix:

| Metric                                                | Type      | Labels                        |
| ----------------------------------------------------- | --------- | ----------------------------- |
| `http_requests_total`                                 | counter   | `method`, `path`, `status`    |
| `http_request_duration_seconds`                       | histogram | `method`, `path`              |
| `kube_api_requests_total`                             | counter   | `verb`, `resource`, `code`    |

Plus the default Go runtime and process collectors.

## Layout

```text
services/gateway/
├── cmd/                    Entry point (main.go)
└── internal/
    ├── api/                REST handlers (namespaces, …)
    ├── buildinfo/          Version metadata injected via ldflags
    ├── config/             Environment parsing
    ├── kube/               client-go factory + impersonation
    ├── logging/            slog logger factory
    ├── metrics/            Prometheus registry
    ├── middleware/         Request ID, access log, recover, CORS, metrics
    ├── oidc/               Verifier + auth middleware + dev pass-through
    └── server/             HTTP server assembly
```

## Milestones

- **M1 — Foundations** ✓ skeleton, healthz/readyz/metrics, graceful shutdown.
- **M2 — Kube + OIDC** ✓ client-go, impersonation, OIDC verifier, real
  Prometheus, first domain endpoint (`/api/v1/namespaces`).
- **M6 — Domain endpoints**: Crossplane resource listing, status aggregation.
- **M7 — CRUD endpoints**: apply / diff / delete with YAML validation.
