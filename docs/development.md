# Developer guide

## Prerequisites

| Tool                                     | Minimum version | Purpose                                       |
| ---------------------------------------- | --------------- | --------------------------------------------- |
| [Go](https://go.dev/)                    | 1.26            | Backend services.                             |
| [Node.js](https://nodejs.org/)           | 22              | Frontend toolchain.                           |
| [pnpm](https://pnpm.io/) (via `corepack`) | 10              | Package manager. `corepack enable` sets up.   |
| [Task](https://taskfile.dev/)            | 3               | Task runner.                                  |
| [Helm](https://helm.sh/)                 | 4               | Chart linting and packaging.                  |
| [Docker](https://www.docker.com/) / [Podman](https://podman.io/) | recent | Container images.              |
| [pre-commit](https://pre-commit.com/)    | 4               | Formatter / linter hooks.                     |
| [kind](https://kind.sigs.k8s.io/) or [k3d](https://k3d.io/) | recent | Local Kubernetes cluster.                   |

## Bootstrap

```bash
task bootstrap
# equivalent to:
#   corepack enable && pnpm install --dir web/ui
#   pre-commit install --install-hooks
#   go work sync
```

## Layout

```
.
├── services/                 Go services
│   ├── gateway/              HTTP API + embedded UI
│   └── auth/                 User CR controller + bootstrap
├── web/ui/                   Vue 3 frontend
├── deploy/helm/              Helm chart
├── pkg/                      Shared Go libraries
├── docs/                     Architecture, deployment, ADRs
└── test/                     Integration & e2e suites
```

## Running locally

```bash
task dev
# starts:
#   gateway       :8080  (LOG_FORMAT=text for readable dev logs)
#   auth          :8081
#   vite dev svr  :5173  (proxies /api to :8080)
```

Open <http://localhost:5173>.

## Testing

```bash
task test            # go test + vitest
task test:coverage   # + coverage reports under {pkg,services/*}/coverage
task test:e2e        # playwright suite in test/e2e (wired in M10)
```

## Linting

```bash
task lint
# runs:
#   golangci-lint run ./...
#   eslint . + stylelint
#   helm lint
#   yamllint
#   markdownlint
```

## Building images

```bash
task images                 # docker build -t crossplane-ui/{gateway,auth}:dev
task images:push REGISTRY=registry.example.com
```

## Working on the chart

```bash
task chart:dep              # helm dependency update
task chart:lint             # helm lint + kubeconform
task chart:template         # renders to /tmp/crossplane-ui.rendered.yaml
```

## Working against a real cluster

```bash
kind create cluster --name crossplane-ui
# Install Crossplane v2 (mandatory dependency).
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm upgrade --install crossplane crossplane-stable/crossplane \
    --create-namespace --namespace crossplane-system --version 2.2.0

# Install this chart.
task chart:dep
helm upgrade --install crossplane-ui deploy/helm/crossplane-ui \
    --create-namespace --namespace crossplane-ui \
    --set gateway.image.repository=docker.io/library/gateway \
    --set gateway.image.tag=dev \
    --set auth.image.repository=docker.io/library/auth \
    --set auth.image.tag=dev

# Load the locally-built images into kind.
kind load docker-image library/gateway:dev library/auth:dev --name crossplane-ui
```

## Commit conventions & releases

See [CONTRIBUTING.md](../CONTRIBUTING.md). TL;DR:

- Conventional commits enforced via `commitlint` pre-commit hook.
- `semantic-release` computes the next version and the changelog from the
  commit history on every push to `main`.
- Container images and the Helm chart are published by CI only.
