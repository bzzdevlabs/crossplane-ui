<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { useAuthStore } from '@/stores/auth';
import { loadGatewayConfig } from '@/services/config';

const { t } = useI18n();
const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

const error = ref<string | null>(null);
const version = ref<string | null>(null);
const busy = ref(false);

loadGatewayConfig()
  .then((cfg) => {
    version.value = cfg.version ?? null;
  })
  .catch(() => {
    // Version badge is cosmetic; swallow the failure and let the real
    // sign-in attempt surface any configuration problem.
  });

async function startLogin(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  error.value = null;
  try {
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/';
    if (auth.devMode) {
      await router.replace(redirect);
      return;
    }
    await auth.signIn(redirect);
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
    busy.value = false;
  }
}
</script>

<template>
  <main class="login">
    <div class="login-card">
      <header>
        <h1>{{ t('app.title') }}</h1>
        <p class="subtitle">{{ t('app.subtitle') }}</p>
      </header>

      <p class="intro">{{ t('auth.login.intro') }}</p>

      <button type="button" class="primary" :disabled="busy" @click="startLogin">
        {{ auth.devMode ? t('auth.login.devContinue') : t('auth.login.submit') }}
      </button>

      <p v-if="error" class="error">{{ error }}</p>

      <p v-if="version" class="version">v{{ version }}</p>
    </div>
  </main>
</template>

<style scoped>
.login {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 2rem;
  background: var(--color-bg);
}

.login-card {
  width: 100%;
  max-width: 380px;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 2rem 2.25rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  box-shadow: 0 10px 30px rgb(0 0 0 / 8%);
}

header {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

h1 {
  margin: 0;
  font-size: 1.25rem;
}

.subtitle {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 0.9rem;
}

.intro {
  margin: 0;
  color: var(--color-text);
  font-size: 0.95rem;
  line-height: 1.4;
}

button.primary {
  padding: 0.75rem 1rem;
  border-radius: 6px;
  border: 1px solid var(--color-accent);
  background: var(--color-accent);
  color: var(--color-on-accent);
  font: inherit;
  font-weight: 500;
  cursor: pointer;
}

button.primary[disabled] {
  opacity: 0.6;
  cursor: progress;
}

.error {
  margin: 0;
  color: var(--color-danger);
  font-size: 0.875rem;
}

.version {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 0.75rem;
  text-align: right;
}
</style>
