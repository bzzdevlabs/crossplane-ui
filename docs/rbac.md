# RBAC recipes

crossplane-ui authorizes users via **Kubernetes RBAC**
(see [ADR-0003](adr/0003-k8s-impersonation-rbac.md)). Users authenticated
through Dex come in as Kubernetes `users` + `groups`; permissions are
expressed via regular `(Cluster)RoleBinding`s.

## Suggested group-to-role mapping

| Group              | Role bound                                     | Purpose                                            |
| ------------------ | ---------------------------------------------- | -------------------------------------------------- |
| `crossplane-ui:admin`  | `cluster-admin`                             | Full control. Bootstrap admin belongs here.        |
| `crossplane-ui:editor` | `crossplane-admin` (shipped by Crossplane)  | Manage Compositions, XRs, MRs, Providers.          |
| `crossplane-ui:viewer` | `view` (shipped by Kubernetes)              | Read-only.                                         |
| `crossplane-ui:ops`    | custom (see below)                          | Restart reconciles, trigger Operations, no CRUD.   |

## Custom `ClusterRole` — platform operator

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: crossplane-ui-ops
rules:
  - apiGroups: ["apiextensions.crossplane.io"]
    resources: ["compositions", "compositeresourcedefinitions"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["pkg.crossplane.io"]
    resources: ["providers", "functions", "configurations"]
    verbs: ["get", "list", "watch", "patch"]
  - apiGroups: ["ops.crossplane.io"]     # v2-only
    resources: ["operations"]
    verbs: ["create", "get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["get", "list", "watch"]
```

Bind it:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crossplane-ui-ops
subjects:
  - kind: Group
    name: crossplane-ui:ops
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: crossplane-ui-ops
  apiGroup: rbac.authorization.k8s.io
```

## Gateway ServiceAccount permissions

The gateway's own ServiceAccount only needs the
[impersonation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation)
verbs:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: crossplane-ui-gateway-impersonator
rules:
  - apiGroups: [""]
    resources: ["users", "groups", "serviceaccounts"]
    verbs: ["impersonate"]
  - apiGroups: ["authentication.k8s.io"]
    resources: ["userextras/*", "uids"]
    verbs: ["impersonate"]
```

The ClusterRole is bound to the gateway's SA by the Helm chart (wired in
M2).

## Auth ServiceAccount permissions

The `auth` controller only needs to:

- CRUD `users.auth.crossplane-ui.io` and `groups.auth.crossplane-ui.io` (our
  CRDs — created in M3).
- `get` / `update` the Dex config ConfigMap.
- `get` / `update` the bootstrap Secret and the per-user hash Secrets.
- `create` / `update` / `patch` `events` for observability.

A narrowly-scoped `Role` in the release namespace is sufficient; no
cluster-wide permissions.

## Auditing

Kubernetes audit logs will record the **impersonated** user, not the gateway
SA. Correlate with gateway JSON logs via the `request_id` attribute (wired
in M2).
