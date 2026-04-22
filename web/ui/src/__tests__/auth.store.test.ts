import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

import { useAuthStore } from '@/stores/auth';
import { resetGatewayConfigCache } from '@/services/config';
import { resetOidcForTests } from '@/services/oidc';

function mockConfig(enabled: boolean): void {
  vi.stubGlobal(
    'fetch',
    vi.fn(() =>
      Promise.resolve(new Response(
        JSON.stringify({
          version: 'test',
          auth: {
            enabled,
            issuerURL: 'http://dex.local/dex',
            clientID: 'crossplane-ui',
            scopes: ['openid'],
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )),
    ),
  );
}

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    resetGatewayConfigCache();
    resetOidcForTests();
    try {
      window.localStorage.clear();
    } catch {
      // jsdom raises SecurityError on opaque origins; ignore.
    }
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('is unauthenticated on creation', () => {
    const auth = useAuthStore();
    expect(auth.isAuthenticated).toBe(false);
    expect(auth.user).toBeNull();
    expect(auth.idToken).toBeNull();
  });

  it('falls back to dev mode when the gateway disables auth', async () => {
    mockConfig(false);
    const auth = useAuthStore();
    await auth.initialise();
    expect(auth.devMode).toBe(true);
    expect(auth.isAuthenticated).toBe(true);
    expect(auth.user?.username).toBe('dev-admin');
  });

  it('stays unauthenticated when the gateway enables auth but no session exists', async () => {
    mockConfig(true);
    const auth = useAuthStore();
    await auth.initialise();
    expect(auth.devMode).toBe(false);
    expect(auth.isAuthenticated).toBe(false);
  });

  it('clear() resets the session', () => {
    const auth = useAuthStore();
    auth.clear();
    expect(auth.user).toBeNull();
    expect(auth.idToken).toBeNull();
    expect(auth.accessToken).toBeNull();
  });
});
