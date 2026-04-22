import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

import { useAuthStore } from '@/stores/auth';

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('is unauthenticated on creation', () => {
    const auth = useAuthStore();
    expect(auth.isAuthenticated).toBe(false);
    expect(auth.user).toBeNull();
    expect(auth.idToken).toBeNull();
  });

  it('transitions to authenticated after setSession', () => {
    const auth = useAuthStore();
    auth.setSession({ username: 'alice', displayName: 'Alice', groups: [] }, 'token-xyz');
    expect(auth.isAuthenticated).toBe(true);
    expect(auth.user?.username).toBe('alice');
    expect(auth.idToken).toBe('token-xyz');
  });

  it('clears the session', () => {
    const auth = useAuthStore();
    auth.setSession({ username: 'alice', displayName: 'Alice', groups: [] }, 'token-xyz');
    auth.clear();
    expect(auth.isAuthenticated).toBe(false);
    expect(auth.user).toBeNull();
    expect(auth.idToken).toBeNull();
  });
});
