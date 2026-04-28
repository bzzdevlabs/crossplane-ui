<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import type { StatusVariant } from '@/resources/registry';

export type StatusFilterValue = 'all' | StatusVariant;

const { t } = useI18n();

defineProps<{
  modelValue: StatusFilterValue;
  options?: readonly StatusFilterValue[];
}>();

defineEmits<{
  (e: 'update:modelValue', value: StatusFilterValue): void;
}>();

const DEFAULT_OPTIONS: readonly StatusFilterValue[] = [
  'all',
  'ready',
  'pending',
  'degraded',
  'errored',
];

function label(value: StatusFilterValue): string {
  if (value === 'all') return t('common.all');
  return t(`status.variants.${value}`);
}
</script>

<template>
  <div class="group" role="tablist">
    <button
      v-for="opt in options ?? DEFAULT_OPTIONS"
      :key="opt"
      type="button"
      role="tab"
      :aria-selected="modelValue === opt"
      :class="{ active: modelValue === opt }"
      @click="$emit('update:modelValue', opt)"
    >
      {{ label(opt) }}
    </button>
  </div>
</template>

<style scoped>
.group {
  display: inline-flex;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  overflow: hidden;
}

button {
  padding: 0.35rem 0.75rem;
  border: 0;
  border-left: 1px solid var(--color-border);
  background: var(--color-surface);
  color: inherit;
  font: inherit;
  font-size: 0.8rem;
  cursor: pointer;
  text-transform: capitalize;
}

button:first-child {
  border-left: 0;
}

button.active {
  background: var(--color-accent);
  color: var(--color-on-accent);
}
</style>
