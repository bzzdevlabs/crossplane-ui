<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { useAuthStore } from '@/stores/auth';

const { t } = useI18n();
const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

const username = ref('');
const password = ref('');
const error = ref<string | null>(null);

async function submit(): Promise<void> {
  error.value = null;
  // Real OIDC flow against Dex is wired in M4. For now the form is inert
  // and only demonstrates the shape of the session set by the real flow.
  if (!username.value || !password.value) {
    error.value = t('auth.login.invalidCredentials');
    return;
  }
  auth.setSession(
    { username: username.value, displayName: username.value, groups: [] },
    'placeholder-token',
  );
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/';
  await router.replace(redirect);
}
</script>

<template>
  <main class="login">
    <form class="login-card" @submit.prevent="submit">
      <h1>{{ t('auth.login.title') }}</h1>

      <label>
        <span>{{ t('auth.login.username') }}</span>
        <input v-model="username" type="text" autocomplete="username" required />
      </label>

      <label>
        <span>{{ t('auth.login.password') }}</span>
        <input
          v-model="password"
          type="password"
          autocomplete="current-password"
          required
        />
      </label>

      <p v-if="error" class="error">{{ error }}</p>

      <button type="submit">{{ t('auth.login.submit') }}</button>

      <button type="button" class="sso" disabled>{{ t('auth.login.sso') }}</button>
    </form>
  </main>
</template>

<style scoped>
.login {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 2rem;
}

.login-card {
  width: 100%;
  max-width: 360px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 2rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.875rem;
}

input {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface-alt);
  font: inherit;
}

button {
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  border: 1px solid var(--color-accent);
  background: var(--color-accent);
  color: var(--color-on-accent);
  font: inherit;
  cursor: pointer;
}

.sso {
  background: transparent;
  color: var(--color-accent);
}

.sso[disabled] {
  opacity: 0.5;
  cursor: not-allowed;
}

.error {
  margin: 0;
  color: var(--color-danger);
  font-size: 0.875rem;
}
</style>
