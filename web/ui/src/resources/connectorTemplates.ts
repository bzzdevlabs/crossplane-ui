import type { ConnectorType } from '@/services/api';

export interface ConnectorFieldBase {
  readonly key: string;
  readonly label: string;
  readonly placeholder?: string;
  readonly helpText?: string;
  readonly required?: boolean;
}

export interface ConnectorStringField extends ConnectorFieldBase {
  readonly kind: 'string';
}

export interface ConnectorBoolField extends ConnectorFieldBase {
  readonly kind: 'bool';
}

export interface ConnectorStringListField extends ConnectorFieldBase {
  readonly kind: 'stringList';
  readonly separator?: 'comma' | 'newline';
}

export interface ConnectorSecretField extends ConnectorFieldBase {
  readonly kind: 'secret';
  // When the user provides a plaintext value, it is written to a Secret and
  // referenced by `secretRefs` pointing at `path`.
  readonly path: string;
  readonly secretKey: string;
}

export type ConnectorField =
  | ConnectorStringField
  | ConnectorBoolField
  | ConnectorStringListField
  | ConnectorSecretField;

export interface ConnectorTemplate {
  readonly type: ConnectorType;
  readonly label: string;
  readonly description: string;
  readonly docsHref?: string;
  readonly fields: readonly ConnectorField[];
  // Defaults applied to .spec.config at creation time.
  readonly defaults: Record<string, unknown>;
}

export const CONNECTOR_TEMPLATES: readonly ConnectorTemplate[] = [
  {
    type: 'ldap',
    label: 'LDAP',
    description: 'Any directory that speaks LDAP v3 (OpenLDAP, FreeIPA, AD).',
    docsHref: 'https://dexidp.io/docs/connectors/ldap/',
    defaults: {
      insecureNoSSL: false,
      insecureSkipVerify: false,
      userSearch: { baseDN: '', username: 'uid', idAttr: 'uid', emailAttr: 'mail', nameAttr: 'cn' },
      groupSearch: { baseDN: '', userAttr: 'DN', groupAttr: 'member', nameAttr: 'cn' },
    },
    fields: [
      { kind: 'string', key: 'host', label: 'Host (host:port)', required: true, placeholder: 'ldap.example.com:636' },
      { kind: 'bool', key: 'insecureNoSSL', label: 'Disable TLS (dev only)' },
      { kind: 'bool', key: 'insecureSkipVerify', label: 'Skip TLS verification' },
      { kind: 'string', key: 'bindDN', label: 'Bind DN', placeholder: 'cn=admin,dc=example,dc=com' },
      { kind: 'secret', key: 'bindPW', label: 'Bind password', path: 'bindPW', secretKey: 'bindPW' },
      { kind: 'string', key: 'userSearch.baseDN', label: 'User search base DN', required: true },
      { kind: 'string', key: 'userSearch.username', label: 'User search username attribute' },
      { kind: 'string', key: 'userSearch.idAttr', label: 'User id attribute' },
      { kind: 'string', key: 'userSearch.emailAttr', label: 'Email attribute' },
      { kind: 'string', key: 'userSearch.nameAttr', label: 'Name attribute' },
      { kind: 'string', key: 'groupSearch.baseDN', label: 'Group search base DN' },
      { kind: 'string', key: 'groupSearch.userAttr', label: 'Group search user attribute' },
      { kind: 'string', key: 'groupSearch.groupAttr', label: 'Group search group attribute' },
      { kind: 'string', key: 'groupSearch.nameAttr', label: 'Group name attribute' },
    ],
  },
  {
    type: 'saml',
    label: 'SAML 2.0',
    description: 'Okta, Azure AD (SAML), ADFS, OneLogin, Shibboleth, Keycloak.',
    docsHref: 'https://dexidp.io/docs/connectors/saml/',
    defaults: {
      redirectURI: '',
      ssoURL: '',
      entityIssuer: '',
      usernameAttr: 'name',
      emailAttr: 'email',
      groupsAttr: 'groups',
    },
    fields: [
      { kind: 'string', key: 'ssoURL', label: 'SSO URL', required: true, placeholder: 'https://idp.example.com/saml/sso' },
      { kind: 'string', key: 'redirectURI', label: 'Redirect URI (ACS)', required: true, helpText: 'Usually https://<dex>/callback.' },
      { kind: 'string', key: 'entityIssuer', label: 'Entity issuer (SP entity ID)' },
      { kind: 'secret', key: 'caData', label: 'IdP CA bundle (PEM)', path: 'caData', secretKey: 'ca.crt' },
      { kind: 'string', key: 'usernameAttr', label: 'Username attribute' },
      { kind: 'string', key: 'emailAttr', label: 'Email attribute' },
      { kind: 'string', key: 'groupsAttr', label: 'Groups attribute' },
    ],
  },
  {
    type: 'github',
    label: 'GitHub',
    description: 'OAuth against github.com or a GitHub Enterprise instance.',
    docsHref: 'https://dexidp.io/docs/connectors/github/',
    defaults: {
      clientID: '',
      redirectURI: '',
      loadAllGroups: false,
      teamNameField: 'slug',
    },
    fields: [
      { kind: 'string', key: 'clientID', label: 'Client ID', required: true },
      { kind: 'secret', key: 'clientSecret', label: 'Client secret', path: 'clientSecret', secretKey: 'clientSecret' },
      { kind: 'string', key: 'redirectURI', label: 'Redirect URI', required: true },
      { kind: 'stringList', key: 'orgs.0.name', label: 'Allowed org(s)', helpText: 'Single org name; leave empty to accept any.' },
      { kind: 'bool', key: 'loadAllGroups', label: 'Load all groups (orgs + teams)' },
      { kind: 'string', key: 'teamNameField', label: 'Team name field (name or slug)' },
    ],
  },
  {
    type: 'google',
    label: 'Google',
    description: 'Google Workspace OIDC.',
    docsHref: 'https://dexidp.io/docs/connectors/google/',
    defaults: {
      clientID: '',
      redirectURI: '',
      hostedDomains: [],
    },
    fields: [
      { kind: 'string', key: 'clientID', label: 'Client ID', required: true },
      { kind: 'secret', key: 'clientSecret', label: 'Client secret', path: 'clientSecret', secretKey: 'clientSecret' },
      { kind: 'string', key: 'redirectURI', label: 'Redirect URI', required: true },
      { kind: 'stringList', key: 'hostedDomains', label: 'Allowed hosted domain(s)', helpText: 'Comma separated.' },
    ],
  },
  {
    type: 'oidc',
    label: 'Generic OIDC',
    description: 'Any OIDC provider (Keycloak, Auth0, Okta OIDC, Azure AD OIDC, ...).',
    docsHref: 'https://dexidp.io/docs/connectors/oidc/',
    defaults: {
      issuer: '',
      clientID: '',
      redirectURI: '',
      insecureSkipEmailVerified: false,
    },
    fields: [
      { kind: 'string', key: 'issuer', label: 'Issuer URL', required: true, placeholder: 'https://accounts.example.com' },
      { kind: 'string', key: 'clientID', label: 'Client ID', required: true },
      { kind: 'secret', key: 'clientSecret', label: 'Client secret', path: 'clientSecret', secretKey: 'clientSecret' },
      { kind: 'string', key: 'redirectURI', label: 'Redirect URI', required: true },
      { kind: 'bool', key: 'insecureSkipEmailVerified', label: 'Accept unverified e-mails' },
    ],
  },
];

export function templateFor(type: ConnectorType): ConnectorTemplate | undefined {
  return CONNECTOR_TEMPLATES.find((t) => t.type === type);
}

export function getAtPath(obj: Record<string, unknown>, path: string): unknown {
  const parts = path.split('.');
  let cur: unknown = obj;
  for (const p of parts) {
    if (cur == null || typeof cur !== 'object') return undefined;
    cur = (cur as Record<string, unknown>)[p];
  }
  return cur;
}

export function setAtPath(obj: Record<string, unknown>, path: string, val: unknown): void {
  const parts = path.split('.');
  let cur: Record<string, unknown> = obj;
  for (let i = 0; i < parts.length - 1; i += 1) {
    const p = parts[i]!;
    const next = cur[p];
    if (next == null || typeof next !== 'object' || Array.isArray(next)) {
      const nm: Record<string, unknown> = {};
      cur[p] = nm;
      cur = nm;
    } else {
      cur = next as Record<string, unknown>;
    }
  }
  cur[parts[parts.length - 1]!] = val;
}
