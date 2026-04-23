<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import type { StatusVariant } from '@/resources/registry';

const props = defineProps<{
  variant: StatusVariant;
  label?: string;
}>();

const { t } = useI18n();

const text = computed(() => props.label ?? t(`status.variants.${props.variant}`));
</script>

<template>
  <span class="pill" :data-variant="variant">
    <span class="dot" aria-hidden="true" />
    {{ text }}
  </span>
</template>

<style scoped>
.pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.15rem 0.6rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 500;
  border: 1px solid transparent;
  line-height: 1.2;
}

.dot {
  width: 0.45rem;
  height: 0.45rem;
  border-radius: 999px;
  background: currentcolor;
}

.pill[data-variant='ready'] {
  color: #1f7a3a;
  border-color: #1f7a3a;
  background: #e6f6ed;
}

.pill[data-variant='degraded'],
.pill[data-variant='errored'] {
  color: #a3341f;
  border-color: #a3341f;
  background: #fdebe7;
}

.pill[data-variant='pending'] {
  color: #8a6100;
  border-color: #c9a227;
  background: #fff4d1;
}

.pill[data-variant='unknown'] {
  color: var(--color-text-muted);
  border-color: var(--color-border);
  background: var(--color-surface-alt);
}

@media (prefers-color-scheme: dark) {
  .pill[data-variant='ready'] {
    background: rgb(31 122 58 / 18%);
  }

  .pill[data-variant='degraded'],
  .pill[data-variant='errored'] {
    background: rgb(163 52 31 / 20%);
  }

  .pill[data-variant='pending'] {
    background: rgb(201 162 39 / 20%);
    color: #e6c14a;
    border-color: #c9a227;
  }
}
</style>
