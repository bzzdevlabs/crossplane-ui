import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import type { User as OidcUser } from 'oidc-client-ts';

import { getUserManager, isAuthEnabled } from '@/services/oidc';

export interface AuthUser {
  readonly subject: string;
  readonly username: string;
  readonly displayName: string;
  readonly email?: string;
  readonly groups: readonly string[];
}

interface Claims {
  readonly sub?: string;
  readonly preferred_username?: string;
  readonly name?: string;
  readonly email?: string;
  readonly groups?: readonly string[];
}

function toAuthUser(u: OidcUser): AuthUser {
  const claims = u.profile as Claims;
  const username = claims.preferred_username ?? claims.email ?? claims.sub ?? 'user';
  return {
    subject: claims.sub ?? '',
    username,
    displayName: claims.name ?? username,
    email: claims.email,
    groups: claims.groups ?? [],
  };
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null);
  const idToken = ref<string | null>(null);
  const accessToken = ref<string | null>(null);
  const ready = ref(false);
  const devMode = ref(false);

  const isAuthenticated = computed(
    () => devMode.value || (user.value !== null && idToken.value !== null),
  );

  function applyOidc(u: OidcUser | null): void {
    if (u === null) {
      user.value = null;
      idToken.value = null;
      accessToken.value = null;
      return;
    }
    user.value = toAuthUser(u);
    idToken.value = u.id_token ?? null;
    accessToken.value = u.access_token ?? null;
  }

  /**
   * Initialise the store from localStorage (if a previous session exists)
   * and subscribe to silent-renew events. In dev pass-through mode the store
   * is put into `devMode`, which bypasses every subsequent guard.
   */
  async function initialise(): Promise<void> {
    if (ready.value) return;
    try {
      const enabled = await isAuthEnabled();
      if (!enabled) {
        devMode.value = true;
        user.value = {
          subject: 'dev',
          username: 'dev-admin',
          displayName: 'dev-admin',
          email: 'dev-admin@local',
          groups: ['system:masters'],
        };
        ready.value = true;
        return;
      }
      const mgr = await getUserManager();
      const existing = await mgr.getUser();
      if (existing && !existing.expired) {
        applyOidc(existing);
      }
      mgr.events.addUserLoaded((u) => applyOidc(u));
      mgr.events.addUserUnloaded(() => applyOidc(null));
      mgr.events.addAccessTokenExpired(() => applyOidc(null));
    } finally {
      ready.value = true;
    }
  }

  async function signIn(redirect?: string): Promise<void> {
    const mgr = await getUserManager();
    await mgr.signinRedirect({ state: redirect ?? '/' });
  }

  async function completeSignIn(): Promise<string> {
    const mgr = await getUserManager();
    const u = await mgr.signinRedirectCallback();
    applyOidc(u);
    const state = typeof u.state === 'string' ? u.state : '/';
    return state;
  }

  async function signOut(): Promise<void> {
    if (devMode.value) {
      // In dev mode the store is synthetic; nothing to clear.
      return;
    }
    const mgr = await getUserManager();
    try {
      await mgr.signoutRedirect({ id_token_hint: idToken.value ?? undefined });
    } catch {
      // Dex may not advertise end_session_endpoint; fall back to local clear.
      await mgr.removeUser();
      applyOidc(null);
    }
  }

  function clear(): void {
    applyOidc(null);
  }

  return {
    user,
    idToken,
    accessToken,
    ready,
    devMode,
    isAuthenticated,
    initialise,
    signIn,
    completeSignIn,
    signOut,
    clear,
  };
});
