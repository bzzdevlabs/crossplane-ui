<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { useAuthStore } from '@/stores/auth';

const { t } = useI18n();
const router = useRouter();
const auth = useAuthStore();
const error = ref<string | null>(null);

onMounted(async () => {
  try {
    const target = await auth.completeSignIn();
    await router.replace(target && target.startsWith('/') ? target : '/');
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  }
});
</script>

<template>
  <main class="callback">
    <p v-if="!error">{{ t('auth.callback.processing') }}</p>
    <div v-else class="error">
      <p>{{ t('auth.callback.failed') }}</p>
      <code>{{ error }}</code>
      <RouterLink :to="{ name: 'login' }">{{ t('auth.callback.retry') }}</RouterLink>
    </div>
  </main>
</template>

<style scoped>
.callback {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 2rem;
  color: var(--color-text-muted);
}

.error {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  align-items: center;
  color: var(--color-danger);
}

.error code {
  font-size: 0.875rem;
  color: var(--color-text);
  background: var(--color-surface-alt);
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  max-width: 48rem;
  overflow-x: auto;
}
</style>
