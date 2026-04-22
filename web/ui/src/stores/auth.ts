import { computed, ref } from 'vue';
import { defineStore } from 'pinia';

export interface AuthUser {
  readonly username: string;
  readonly displayName: string;
  readonly groups: readonly string[];
  readonly email?: string;
}

/**
 * useAuthStore holds the currently authenticated user and the ID token the
 * gateway must forward to the Kubernetes API for impersonation.
 *
 * The real OIDC flow against Dex is wired in milestone M4. Until then the
 * store only exposes the shape the rest of the app will consume.
 */
export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null);
  const idToken = ref<string | null>(null);

  const isAuthenticated = computed(() => user.value !== null && idToken.value !== null);

  function setSession(nextUser: AuthUser, token: string): void {
    user.value = nextUser;
    idToken.value = token;
  }

  function clear(): void {
    user.value = null;
    idToken.value = null;
  }

  return { user, idToken, isAuthenticated, setSession, clear };
});
