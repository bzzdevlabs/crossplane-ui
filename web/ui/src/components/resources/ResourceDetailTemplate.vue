<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import StatusPill from '@/components/ui/StatusPill.vue';
import ActionMenu, { type ActionItem } from '@/components/ui/ActionMenu.vue';
import BreadcrumbBar, { type Crumb } from '@/components/ui/BreadcrumbBar.vue';
import type { StatusVariant } from '@/resources/registry';

defineProps<{
  title: string;
  kind?: string;
  metaParts?: readonly string[];
  status?: StatusVariant;
  breadcrumbs?: readonly Crumb[];
  saving?: boolean;
  canApply?: boolean;
  overflowItems?: readonly ActionItem[];
}>();

defineEmits<{
  (e: 'refresh'): void;
  (e: 'delete'): void;
  (e: 'apply'): void;
  (e: 'overflow', id: string): void;
}>();

const { t } = useI18n();
</script>

<template>
  <section class="detail">
    <BreadcrumbBar v-if="breadcrumbs && breadcrumbs.length" :items="breadcrumbs" />

    <header class="page-header">
      <div class="title-block">
        <div class="title-row">
          <h1>{{ title }}</h1>
          <StatusPill v-if="status" :variant="status" />
        </div>
        <p v-if="kind || (metaParts && metaParts.length)" class="muted">
          <span v-if="kind">{{ kind }}</span>
          <template v-for="(part, idx) in metaParts ?? []" :key="idx">
            <span class="dot">·</span><span>{{ part }}</span>
          </template>
        </p>
      </div>
      <div class="actions">
        <slot name="actions">
          <button type="button" :disabled="saving" @click="$emit('refresh')">
            {{ t('common.refresh') }}
          </button>
          <button type="button" class="danger" :disabled="saving" @click="$emit('delete')">
            {{ t('common.delete') }}
          </button>
          <button
            type="button"
            class="primary"
            :disabled="saving || canApply === false"
            @click="$emit('apply')"
          >
            {{ saving ? t('resource.saving') : t('resource.apply') }}
          </button>
          <ActionMenu
            v-if="overflowItems && overflowItems.length"
            :items="overflowItems"
            @select="(id) => $emit('overflow', id)"
          />
        </slot>
      </div>
    </header>

    <slot />
  </section>
</template>

<style scoped>
.detail {
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

.title-block {
  min-width: 0;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

h1 {
  margin: 0;
  font-size: 1.35rem;
}

.muted {
  margin: 0.2rem 0 0;
  color: var(--color-text-muted);
  font-size: 0.9rem;
  display: inline-flex;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.dot {
  opacity: 0.5;
}

.actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.actions button {
  padding: 0.4rem 0.9rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.actions button[disabled] {
  opacity: 0.5;
  cursor: not-allowed;
}

.actions .primary {
  border-color: var(--color-accent);
  background: var(--color-accent);
  color: var(--color-on-accent);
}

.actions .danger {
  border-color: var(--color-danger);
  color: var(--color-danger);
  background: transparent;
}
</style>
