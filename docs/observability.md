# Observability

Every service exposes three operational endpoints and structured JSON logs.

## Endpoints

| Endpoint   | Port       | Format                         | Purpose                               |
| ---------- | ---------- | ------------------------------ | ------------------------------------- |
| `/healthz` | `:8080/81` | `text/plain`                   | Liveness. Always returns 200 if up.   |
| `/readyz`  | `:8080/81` | `text/plain`                   | Readiness. 503 during startup probes. |
| `/metrics` | `:8080/81` | Prometheus text exposition.    | Scrape target.                        |

## Logs

- Encoder: `log/slog` JSON handler (`LOG_FORMAT=text` for dev).
- Keys are `snake_case` (enforced by `sloglint`).
- Required attributes on every record: `service`, `version`, and for HTTP:
  `method`, `path`, `status`, `duration`, `remote_addr`.

Correlation IDs, upstream span IDs and OTEL propagation are wired in the
`observability` ADR (planned as ADR-0006).

## Metrics wired (per milestone)

| Milestone | Metric (prefix `crossplane_ui_…`)                                            | Status  |
| --------- | ---------------------------------------------------------------------------- | ------- |
| M2        | `gateway_http_requests_total` ({method,path,status})                         | ✓ live  |
| M2        | `gateway_http_request_duration_seconds` ({method,path})                      | ✓ live  |
| M2        | `gateway_kube_api_requests_total` ({verb,resource,code})                     | ✓ live  |
| M3        | `auth_reconciler_loop_duration_seconds`                                      | planned |
| M3        | `auth_users_total`                                                           | planned |
| M6        | `gateway_crossplane_resources_total` ({kind,ready,synced})                   | planned |

## Dashboards

A Grafana dashboard JSON will be committed under `deploy/grafana/` once the
metric shape stabilises (end of M6).

## Alerts

A `PrometheusRule` sidecar chart will ship alongside v1.0.0 with the usual
SRE staples (latency SLO burn, error-rate burn, pod restarts, missing
scrape targets).
