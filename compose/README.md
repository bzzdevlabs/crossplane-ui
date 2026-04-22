# Local dev stack

Single-command local environment with:

- a [k3d](https://k3d.io/) cluster named `crossplane-ui` running
  Kubernetes 1.31 (k3d default);
- **Crossplane v2.2** pre-installed in the cluster;
- `dex` / `auth` / `gateway` / `ui` running as containers in the same Docker
  network as the cluster (`k3d-crossplane-ui`), so they can reach the
  Kubernetes API without additional wiring.

## Prerequisites

- [Docker](https://www.docker.com/) 24+ with Compose v2.
- [k3d](https://k3d.io/) 5.8+.
- [Helm](https://helm.sh/) 4+.
- [Task](https://taskfile.dev/) 3+.

## Quick start

```bash
task dev:up            # creates k3d + installs Crossplane + docker compose up
# → UI      http://localhost:5173
# → API     http://localhost:8080
# → Dex     http://localhost:5556/dex
# → K8s     kubectl cluster-info  (after running task dev:kubeconfig)

task dev:down          # stops compose services (keeps the cluster running)
task dev:reset         # destroys everything (cluster + compose + volumes)
task dev:logs          # follows compose logs
task dev:kubeconfig    # prints the `export KUBECONFIG=...` line for the host
```

## What runs where

| Component  | Where                             | Port(s)            | Purpose                                    |
| ---------- | --------------------------------- | ------------------ | ------------------------------------------ |
| k3s        | k3d node container                | 6443 (via k3d LB)  | Kubernetes API + Crossplane controllers.   |
| Crossplane | k3s, `crossplane-system` ns       | n/a                | Provides the CRDs the UI will list/manage. |
| `dex`      | compose service                   | 5556               | OIDC IdP with two local users.             |
| `auth`     | compose service (golang:1.26)      | 8081               | Local-user controller (scaffold — M1).     |
| `gateway`  | compose service (golang:1.26)      | 8080               | HTTP API (scaffold — M1).                  |
| `ui`       | compose service (node:22 + pnpm)   | 5173 (Vite dev)    | Hot-reloading Vue app.                     |

## Network layout

```
┌─────────────── docker network "k3d-crossplane-ui" ───────────────┐
│                                                                  │
│  ┌──────────────────────┐   ┌──────────┐   ┌──────────┐          │
│  │ k3d-*-server-0       │   │ gateway  │──▶│   auth   │          │
│  │ (k3s API 6443)       │◀──│  :8080   │   │   :8081  │          │
│  └──────────────────────┘   └────┬─────┘   └──────────┘          │
│                                  │                               │
│                                  ▼                               │
│                             ┌────────┐                           │
│                             │  dex   │                           │
│                             │ :5556  │                           │
│                             └────────┘                           │
│                                                                  │
│                             ┌────────┐                           │
│                             │   ui   │  (Vite proxies /api →     │
│                             │ :5173  │    http://gateway:8080)   │
│                             └────────┘                           │
└──────────────────────────────────────────────────────────────────┘
           ▲             ▲           ▲            ▲
           │             │           │            │
     localhost:5173 localhost:8080 localhost:8081 localhost:5556
        (browser)
```

## Default credentials (dev only)

Configured in [`dex/config.yaml`](dex/config.yaml):

| Username                       | Password   |
| ------------------------------ | ---------- |
| `admin@crossplane-ui.local`    | `password` |
| `viewer@crossplane-ui.local`   | `password` |

OIDC client: `crossplane-ui` / `dev-secret-change-me`.

> Full OIDC login flow is wired in milestone **M4**. Until then the UI's
> login page uses a stub that accepts any credentials.

## Inspecting the cluster from the host

```bash
export KUBECONFIG=$(k3d kubeconfig write crossplane-ui)
kubectl get providers -A                     # Crossplane providers
kubectl -n crossplane-system get pods        # Crossplane core
kubectl api-resources | grep crossplane      # all Crossplane CRDs
```

## Deploying a sample Crossplane composition

Once milestones M6+ are merged the UI will surface any Composition /
Composite Resource in the cluster. You can seed sample resources with:

```bash
kubectl apply -f https://raw.githubusercontent.com/crossplane/docs/v2.2/content/v2.2/get-started/apis.yaml
```

## Troubleshooting

### The Go services take forever to start

The first run downloads the full Go module cache inside the compose
`go-mod` volume. Subsequent starts reuse the cache and boot in a few
seconds.

### Vite cannot reach the gateway

Ensure the `gateway` container is healthy (`docker compose ps`). The Vite
proxy target is set via `VITE_PROXY_TARGET=http://gateway:8080` — this
relies on Docker's internal DNS, not on `localhost`.

### I get TLS errors from the Go services

The generated `.kubeconfig` sets `insecure-skip-tls-verify: true` because
k3d's API certificate does not include our internal DNS name
(`k3d-crossplane-ui-server-0`) as a SAN. This is acceptable for
development only.

### I want to use my own cluster instead of k3d

Delete the `networks.default.external` block in `docker-compose.yml` and
point `KUBECONFIG` at your own kubeconfig (mounted into the `auth` and
`gateway` services). You will also need to adjust the `OIDC_ISSUER_URL`
so the gateway can reach Dex from outside the compose network.

## Alternative: all-in-one compose (no k3d)

Prefer a single `docker compose up` with no external tools? See
[`docs/development.md`](../docs/development.md#all-in-one-compose) for a
variant that embeds `rancher/k3s` directly inside compose. It trades a
slower boot and higher resource usage for a zero-dependency run.
