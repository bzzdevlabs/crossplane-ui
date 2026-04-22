# crossplane-ui Helm chart

Deploys the `gateway` and `auth` services plus an embedded [Dex](https://dexidp.io/)
identity provider into a Kubernetes cluster that already runs Crossplane v2.

## Installation

```bash
helm repo add dex https://charts.dexidp.io   # required for the Dex dependency
helm dependency update deploy/helm/crossplane-ui
helm upgrade --install crossplane-ui deploy/helm/crossplane-ui \
    --namespace crossplane-ui --create-namespace \
    --values deploy/helm/crossplane-ui/values.yaml
```

The chart's `values.yaml` is fully annotated; a JSON Schema
(`values.schema.json`) validates the inputs at install time.

## Bootstrap administrator

On first install, a `Secret` named `<release>-bootstrap-admin` is created with
the bootstrap admin credentials. If `.Values.auth.bootstrapAdmin.password` is
empty, a random 24-character alphanumeric password is generated. Read it with:

```bash
kubectl -n crossplane-ui get secret crossplane-ui-bootstrap-admin \
    -o jsonpath='{.data.password}' | base64 -d ; echo
```

The chart marks the Secret with `helm.sh/resource-policy: keep` so that the
password survives upgrades; the `auth` controller becomes the authoritative
owner after bootstrap.

## Dex subchart

`dependencies` in `Chart.yaml` pulls in the upstream
[Dex](https://github.com/dexidp/helm-charts) chart. Override its values under
the `dex:` key in our `values.yaml`. Connector configuration (LDAP / SAML /
GitHub / Google / …) is templated there.

## What is installed (at milestone M1)

| Resource               | Count | Purpose                                          |
| ---------------------- | ----- | ------------------------------------------------ |
| `Deployment/gateway`   | 1     | HTTP entrypoint (2 replicas by default).         |
| `Deployment/auth`      | 1     | Controller + bootstrap (single leader).          |
| `Service` × 2          | 2     | ClusterIP for each deployment.                   |
| `ServiceAccount` × 2   | 2     | One per component (impersonation wired in M2).   |
| `Secret/bootstrap`     | 1     | First-install admin credentials.                 |
| `NetworkPolicy` × 3    | 3     | Deny-all + per-component allowances.             |
| `Ingress`              | 0/1   | Optional (`ingress.enabled`).                    |
| `ServiceMonitor`       | 0/1   | Optional (`serviceMonitor.enabled`).             |
| Dex subchart           | —     | OIDC IdP fronting all authentication protocols.  |

RBAC (Roles / RoleBindings for the `auth` controller and for the `gateway`
impersonation) is wired in milestones M2 and M3.

## Development

```bash
task chart:lint     # helm lint + kubeconform schema validation
task chart:template # render the chart to stdout for manual inspection
task chart:test     # helm unittest (wired in M9)
```
