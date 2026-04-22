<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { RouterLink } from 'vue-router';

const { t } = useI18n();
</script>

<template>
  <main class="not-found">
    <h1>{{ t('notFound.heading') }}</h1>
    <p>{{ t('notFound.message') }}</p>
    <RouterLink :to="{ name: 'home' }">{{ t('notFound.backHome') }}</RouterLink>
  </main>
</template>

<style scoped>
.not-found {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 2rem;
  text-align: center;
  gap: 1rem;
}
</style>
