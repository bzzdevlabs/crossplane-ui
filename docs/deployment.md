# Deployment guide

## Prerequisites

- Kubernetes **1.33+** (matches the Crossplane v2 support matrix).
- Crossplane v2 already installed in the target cluster.
- [Helm 4+](https://helm.sh/).
- An `Ingress` controller (nginx, traefik, …) if you want external access.
- [cert-manager](https://cert-manager.io/) if you want automatic TLS.
- A Prometheus operator if you want to scrape our `ServiceMonitor`.

## Install

The chart ships through three channels — pick whichever fits your
workflow (all three publish identical artifacts; see ADR-0009 + ADR-0010
for the rationale).

### Channel A — Helm repository (recommended)

```bash
helm repo add crossplane-ui https://bzzdevlabs.github.io/crossplane-ui
helm repo update
helm upgrade --install crossplane-ui crossplane-ui/crossplane-ui \
    --version <X.Y.Z> \
    --namespace crossplane-ui --create-namespace \
    --set ingress.enabled=true \
    --set ingress.className=nginx \
    --set ingress.hosts[0].host=crossplane-ui.example.com \
    --set ingress.hosts[0].paths[0].path=/ \
    --set ingress.hosts[0].paths[0].pathType=Prefix
```

### Channel B — OCI (no `helm repo add`)

```bash
helm upgrade --install crossplane-ui \
    oci://ghcr.io/bzzdevlabs/crossplane-ui/charts/crossplane-ui \
    --version <X.Y.Z> \
    --namespace crossplane-ui --create-namespace
```

### Channel C — From source (development)

```bash
helm repo add dex https://charts.dexidp.io
helm dependency update deploy/helm/crossplane-ui

helm upgrade --install crossplane-ui deploy/helm/crossplane-ui \
    --namespace crossplane-ui --create-namespace
```

### Bootstrap administrator

On first install the chart generates (or accepts) a password for the bootstrap
admin and stores it in the `<release>-bootstrap-admin` Secret.

```bash
kubectl -n crossplane-ui get secret crossplane-ui-bootstrap-admin \
    -o jsonpath='{.data.password}' | base64 -d ; echo
```

Rotate the password through the UI after the first login. The Secret is
preserved on upgrades (`helm.sh/resource-policy: keep`).

### External identity providers (LDAP, SAML, OIDC, OAuth)

Override the `dex.config.connectors` key in your `values.yaml`. See
[docs/authentication.md](authentication.md) for per-provider examples.

## Upgrade

```bash
helm upgrade crossplane-ui deploy/helm/crossplane-ui \
    --namespace crossplane-ui \
    --values values-prod.yaml
```

The `auth` Deployment uses `strategy: Recreate` because it is a single-leader
controller. Expect a short (< 30 s) gap in reconciliation during upgrades.

## Uninstall

```bash
helm uninstall crossplane-ui -n crossplane-ui
kubectl delete namespace crossplane-ui
```

The bootstrap Secret and any `User` CRs are not deleted automatically (by
design) — remove them manually if you want a clean slate.

## Hardening checklist

- [ ] TLS on the public Ingress (cert-manager).
- [ ] `NetworkPolicy` enabled (`networkPolicy.enabled=true`, default).
- [ ] Rotate the bootstrap admin password through the UI.
- [ ] Set a non-default `oidc.clientSecret`.
- [ ] Lock the chart to a specific image tag (no `latest`).
- [ ] Audit the RBAC rules granted to the `auth` ServiceAccount.
- [ ] Enable the `ServiceMonitor` and wire alerting rules (see
      [docs/observability.md](observability.md)).

## Troubleshooting

### The gateway pod crash-loops

Inspect logs:

```bash
kubectl -n crossplane-ui logs deploy/crossplane-ui-gateway
```

Common causes:

- `OIDC_ISSUER_URL` is unreachable → Dex is not ready yet; the gateway exits
  the OIDC verifier setup. Wait for Dex to be ready, or set
  `dex.enabled=false` and point `oidc.issuerURL` at an external IdP.
- `HTTP_ADDR` conflicts with another container — the default is `:8080`.

### `helm dep update` fails to fetch `dex`

Check DNS and proxy settings; if `charts.dexidp.io` is not reachable, mirror
the Dex chart into an internal Helm registry and override the
`dependencies[0].repository` in `Chart.yaml`.
