<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<template>
  <section class="placeholder">
    <h1>{{ t('products.settings.label') }}</h1>
    <p class="muted">{{ t('products.settings.subtitle') }}</p>
    <p class="coming-soon">Coming later.</p>
  </section>
</template>

<style scoped>
.placeholder {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

h1 {
  margin: 0;
  font-size: 1.5rem;
}

.muted {
  margin: 0;
  color: var(--color-text-muted);
}

.coming-soon {
  margin-top: 1rem;
  color: var(--color-text-muted);
  font-style: italic;
}
</style>
