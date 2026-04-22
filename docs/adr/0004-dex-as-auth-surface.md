# 0004. Dex is the only authentication surface

- Date: 2026-04-22
- Status: Accepted
- Deciders: @qgerard
- Tags: auth

## Context

The product needs to authenticate users via:

- Local users (bootstrap admin plus user-managed accounts).
- LDAP / Active Directory.
- SAML 2.0.
- Upstream OIDC (Keycloak, Google, Azure AD, …).
- OAuth (GitHub, GitLab, …).

Implementing each protocol ourselves would multiply attack surface and is
outside the maintainer's scope.

## Decision

We will embed [Dex](https://dexidp.io/) as a required sub-chart. All five
authentication modalities become Dex **connectors**. The gateway is a
confidential OIDC client of Dex; it never speaks LDAP, SAML or OAuth
directly. Local users are provisioned by the `auth` service into Dex's
`passwordDB`.

## Consequences

Positive:

- **Reduced attack surface**: we own one OIDC relying-party, not N protocol
  implementations.
- Dex is a small (~30 MB), pure-Go static binary, matching our "lightweight
  microservice" goal.
- Connectors can be added and rotated without touching crossplane-ui code.
- Widely used in the K8s ecosystem (ArgoCD, Gitea, Gardener) → mature and
  well-tested.

Negative:

- One extra Pod to operate.
- Migrating away from Dex later would mean re-implementing the connector
  abstraction.

## Alternatives considered

- **Keycloak** — rejected: Java, Postgres dependency, ~300 MB image, more
  than we need.
- **In-house Go implementation with `go-oidc` + `go-ldap`** — rejected:
  SAML alone would be weeks of work and very risky to implement correctly.
- **External IdP only (no local users)** — rejected: some deployments need
  to work offline / air-gapped with a bootstrap admin.

## References

- <https://dexidp.io/>
- <https://github.com/dexidp/dex>
