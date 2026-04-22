import { UserManager, WebStorageStateStore, type UserManagerSettings } from 'oidc-client-ts';

import { loadGatewayConfig, type AuthConfig } from './config';

let manager: UserManager | null = null;
let active: AuthConfig | null = null;

/**
 * Builds (or returns the cached) UserManager for the current gateway
 * configuration. The SPA is a public client: authorization-code flow with
 * PKCE, no client secret embedded. Redirect and silent-renew URIs are
 * derived from `window.location.origin` so the same bundle works under any
 * hostname.
 */
export async function getUserManager(): Promise<UserManager> {
  if (manager) {
    return manager;
  }
  const cfg = await loadGatewayConfig();
  if (!cfg.auth.enabled) {
    throw new Error('oidc disabled (gateway runs in dev pass-through mode)');
  }
  if (!cfg.auth.issuerURL || !cfg.auth.clientID) {
    throw new Error('oidc config missing issuerURL or clientID');
  }

  const settings: UserManagerSettings = {
    authority: cfg.auth.issuerURL,
    client_id: cfg.auth.clientID,
    redirect_uri: `${window.location.origin}/auth/callback`,
    post_logout_redirect_uri: `${window.location.origin}/login`,
    response_type: 'code',
    scope: (cfg.auth.scopes ?? ['openid', 'profile', 'email', 'groups']).join(' '),
    loadUserInfo: false,
    userStore: new WebStorageStateStore({ store: window.localStorage }),
    stateStore: new WebStorageStateStore({ store: window.localStorage }),
    automaticSilentRenew: true,
  };

  manager = new UserManager(settings);
  active = cfg.auth;
  return manager;
}

/**
 * Returns whether OIDC is enabled without building a UserManager (used by
 * router guards before the login view is reached).
 */
export async function isAuthEnabled(): Promise<boolean> {
  const cfg = await loadGatewayConfig();
  return cfg.auth.enabled;
}

/** Reset cached state (tests only). */
export function resetOidcForTests(): void {
  manager = null;
  active = null;
}

export function activeAuthConfig(): AuthConfig | null {
  return active;
}
