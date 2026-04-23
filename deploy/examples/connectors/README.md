# Connector examples

Sample `auth.crossplane-ui.io/v1alpha1` Connector manifests for the five
identity provider types the UI ships forms for.

| File | Provider | Notes |
|------|----------|-------|
| [ldap.yaml](ldap.yaml) | LDAP | OpenLDAP / FreeIPA / AD (LDAP v3) |
| [saml.yaml](saml.yaml) | SAML 2.0 | Okta, Azure AD SAML, ADFS, Keycloak SAML, ... |
| [github.yaml](github.yaml) | GitHub OAuth | github.com or GitHub Enterprise |
| [google.yaml](google.yaml) | Google Workspace | Google OIDC |
| [oidc.yaml](oidc.yaml) | Generic OIDC | Keycloak, Auth0, Okta OIDC, Azure AD OIDC |

Every example carries a companion `Secret` for values that must not live
inside the CR (client secrets, LDAP bind passwords, SAML CA bundles).
`spec.secretRefs` tells the auth controller which JSON path in
`spec.config` to splice each secret value into at project time.

Apply one of the examples:

```bash
kubectl apply -f deploy/examples/connectors/github.yaml
kubectl get connectors
kubectl -n crossplane-ui get connectors.dex.coreos.com
```

The Dex connector object appears in the release namespace once the
controller has reconciled. Dex picks it up at the next request; no
restart is required.
