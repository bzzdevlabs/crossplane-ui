<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { RouteLocationRaw } from 'vue-router';
import { Plus, RotateCw } from 'lucide-vue-next';

defineProps<{
  title: string;
  subtitle?: string;
  createTo?: RouteLocationRaw;
  loading?: boolean;
  error?: string | null;
  count?: number;
}>();

defineEmits<{
  (e: 'refresh'): void;
}>();

const { t } = useI18n();
</script>

<template>
  <section class="list-view">
    <header class="page-header">
      <div>
        <h1>
          {{ title }}
          <span v-if="count !== undefined" class="count">{{ count }}</span>
        </h1>
        <p v-if="subtitle" class="muted">{{ subtitle }}</p>
      </div>
      <div class="actions">
        <slot name="toolbar" />
        <button
          v-if="!loading"
          type="button"
          class="refresh"
          @click="$emit('refresh')"
        >
          <RotateCw :size="14" aria-hidden="true" />
          {{ t('common.refresh') }}
        </button>
        <RouterLink v-if="createTo" :to="createTo" class="create">
          <Plus :size="14" aria-hidden="true" />
          {{ t('resource.list.create') }}
        </RouterLink>
      </div>
    </header>

    <p v-if="loading" class="muted">{{ t('home.loading') }}</p>
    <p v-else-if="error" class="error">{{ t('common.error') }}: {{ error }}</p>
    <slot v-else />
  </section>
</template>

<style scoped>
.list-view {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 1rem;
  flex-wrap: wrap;
}

h1 {
  margin: 0;
  font-size: 1.5rem;
}

.count {
  color: var(--color-text-muted);
  font-size: 1rem;
  font-weight: 400;
  margin-left: 0.35rem;
}

.muted {
  color: var(--color-text-muted);
  margin: 0.25rem 0 0;
}

.error {
  color: var(--color-danger);
}

.actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.refresh,
.create {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.4rem 0.9rem;
  border-radius: 6px;
  font: inherit;
  font-size: 0.9rem;
  cursor: pointer;
  text-decoration: none;
}

.refresh {
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: inherit;
}

.create {
  border: 1px solid var(--color-accent);
  background: var(--color-accent);
  color: var(--color-on-accent);
}
</style>
