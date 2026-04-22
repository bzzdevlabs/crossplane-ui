# 0003. Kubernetes user impersonation for RBAC

- Date: 2026-04-22
- Status: Accepted
- Deciders: @qgerard
- Tags: auth, security

## Context

We need to enforce authorization in the UI. Two broad options exist:

1. The gateway holds a high-privilege `ServiceAccount`, performs actions on
   behalf of the user, and enforces RBAC in application code based on the
   user's role.
2. The gateway forwards the user's identity to the Kubernetes API using
   [user impersonation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation)
   and lets the Kubernetes RBAC sub-system authorize the call.

## Decision

We will use **Kubernetes user impersonation**. The gateway's ServiceAccount
is granted the `impersonate` verb on `users`, `groups` and
`serviceaccounts`. For every incoming request we attach:

- `Impersonate-User: <ID-token 'sub' claim>`
- `Impersonate-Group: <every value in the 'groups' claim>`
- `Impersonate-Extra-oidc-email: <email>` (for audit correlation)

Authorization decisions are then made by the Kubernetes API server against
regular `Role` / `ClusterRole` bindings — exactly the same surface
cluster-admins already curate for their human users.

## Consequences

Positive:

- **Zero secondary RBAC system** to maintain — we use the one administrators
  already know and trust.
- **Correct audit trail**: Kubernetes audit events record the actual user,
  not a shared ServiceAccount.
- **Defense in depth**: if the gateway is compromised, the attacker is still
  limited by the `impersonate` verb's scope, not cluster-admin.

Negative:

- Every call takes an extra header round-trip — negligible.
- Operators must understand the `impersonate` verb and how to scope it (via
  `resourceNames` on the ClusterRole).
- We still need a minimal UI role model ("platform-admin", "viewer", …) but
  these become plain Kubernetes `Group` bindings rather than CRDs.

## Alternatives considered

- **Application-level RBAC**. Rejected because it duplicates the concern
  Kubernetes already solves, complicates audit, and is the usual source of
  privilege-escalation bugs in admin UIs.
- **Per-user ServiceAccount + token exchange**. More expensive (one SA per
  user, one Secret per SA) and Kubernetes actively discourages the pattern.

## References

- <https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation>
- ArgoCD's similar design:
  <https://argo-cd.readthedocs.io/en/stable/operator-manual/rbac/>
