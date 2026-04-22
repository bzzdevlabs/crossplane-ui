<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<template>
  <section>
    <h1>{{ t('home.heading') }}</h1>
    <p class="muted">{{ t('home.empty') }}</p>
    <!--
      Cards/tiles dashboard is wired in milestone M6.
      Each tile shows a Crossplane resource with ready/synced status badges.
    -->
  </section>
</template>

<style scoped>
h1 {
  font-size: 1.5rem;
  margin: 0 0 0.5rem;
}

.muted {
  color: var(--color-text-muted);
}
</style>
