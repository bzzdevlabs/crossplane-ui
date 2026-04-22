# Authentication

`crossplane-ui` uses [Dex IdP](https://dexidp.io/) as its single identity
provider, and the `gateway` service as an OIDC client of Dex. Local users live
in `User` CRDs managed by the `auth` service; all other identity sources
(LDAP, SAML, upstream OIDC, OAuth) are Dex **connectors**.

## Architecture

```
 ┌───────────┐           ┌─────────┐           ┌────────────────┐
 │  Browser  │─── OIDC──▶│   Dex   │── LDAP ──▶│  AD / OpenLDAP │
 │           │  redirect │         │── SAML ──▶│  Azure / ADFS  │
 │           │           │         │── OIDC ──▶│  Google        │
 │           │           │         │── OAuth ▶│  GitHub / GitLab│
 │           │           │         │── static ▶│  auth service  │
 └───────────┘           └─────────┘           └────────────────┘
        │                     ▲
        │                     │
        ▼                     │
 ┌───────────┐     ID token   │
 │  gateway  │◀────────────── │
 │  (OIDC    │
 │  client)  │
 └───────────┘
```

## Local users

- Stored as `User` CRs (apiVersion `auth.crossplane-ui.io/v1alpha1`).
- The bcrypt hash of the password lives in a companion `Secret` owned by the
  `User` CR; only the `auth` ServiceAccount can read it.
- The `auth` controller projects these users into a ConfigMap consumed by Dex
  (`DEX_CONFIGMAP_NAME`); Dex reloads automatically.

## Bootstrap administrator

- The Helm chart provisions a `Secret` named `<release>-bootstrap-admin` on
  first install.
- The `auth` controller watches this Secret: on startup it:
  1. Creates a `User` CR with the username from the Secret.
  2. Writes the bcrypt hash of the Secret's `password` into the companion
     hash Secret.
  3. Grants the `User` the `crossplane-ui:admin` group.
- The `Secret` is kept across upgrades (`helm.sh/resource-policy: keep`);
  rotating the admin password via the UI updates both the Secret and the hash.

## Configuring connectors

All samples below go under the `dex.config.connectors:` key of `values.yaml`.
The upstream Dex documentation lists the full set of options:
<https://dexidp.io/docs/connectors/>.

### LDAP

```yaml
dex:
  config:
    connectors:
      - type: ldap
        id: corp-ldap
        name: Corporate LDAP
        config:
          host: ldap.corp.example.com:636
          insecureNoSSL: false
          bindDN: uid=dex,ou=service,dc=corp,dc=example,dc=com
          bindPW: $LDAP_BIND_PASSWORD   # read from secretEnv
          usernamePrompt: Corporate username
          userSearch:
            baseDN: ou=people,dc=corp,dc=example,dc=com
            filter: "(objectClass=person)"
            username: uid
            idAttr: uid
            emailAttr: mail
            nameAttr: cn
          groupSearch:
            baseDN: ou=groups,dc=corp,dc=example,dc=com
            filter: "(objectClass=group)"
            userMatchers:
              - userAttr: DN
                groupAttr: member
            nameAttr: cn
```

### SAML 2.0 (Azure AD, ADFS, Okta, …)

```yaml
dex:
  config:
    connectors:
      - type: saml
        id: okta-saml
        name: Okta
        config:
          ssoURL: https://<tenant>.okta.com/app/<app_id>/sso/saml
          ca: /etc/dex/saml-ca.pem
          entityIssuer: https://crossplane-ui.example.com/dex/callback
          redirectURI: https://crossplane-ui.example.com/dex/callback
          usernameAttr: email
          emailAttr: email
          groupsAttr: groups
```

### Upstream OIDC (Keycloak, Google)

```yaml
dex:
  config:
    connectors:
      - type: oidc
        id: google
        name: Google
        config:
          issuer: https://accounts.google.com
          clientID: $GOOGLE_CLIENT_ID
          clientSecret: $GOOGLE_CLIENT_SECRET
          redirectURI: https://crossplane-ui.example.com/dex/callback
          scopes:
            - openid
            - email
            - profile
```

### OAuth (GitHub, GitLab)

```yaml
dex:
  config:
    connectors:
      - type: github
        id: github
        name: GitHub
        config:
          clientID: $GITHUB_CLIENT_ID
          clientSecret: $GITHUB_CLIENT_SECRET
          redirectURI: https://crossplane-ui.example.com/dex/callback
          orgs:
            - name: my-org
```

## RBAC

Dex issues an ID token containing the user's `sub` (subject) and `groups`
claims. The gateway forwards those as
`Impersonate-User` / `Impersonate-Group` headers to the Kubernetes API.
Authorization is therefore expressed as plain Kubernetes RBAC, for example:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crossplane-ui-admins
subjects:
  - kind: Group
    name: crossplane-ui:admin
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: crossplane-admin    # bundled by Crossplane itself
  apiGroup: rbac.authorization.k8s.io
```

See [docs/rbac.md](rbac.md) for the full recommended set of ClusterRoles.
