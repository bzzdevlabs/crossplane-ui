<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { Search } from 'lucide-vue-next';

defineProps<{
  modelValue: string;
  placeholder?: string;
}>();

defineEmits<{
  (e: 'update:modelValue', value: string): void;
}>();

const { t } = useI18n();
</script>

<template>
  <label class="filter">
    <span class="sr-only">{{ t('common.filter') }}</span>
    <Search class="icon" :size="14" aria-hidden="true" />
    <input
      type="search"
      :value="modelValue"
      :placeholder="placeholder ?? t('common.filter')"
      @input="(e) => $emit('update:modelValue', (e.target as HTMLInputElement).value)"
    />
  </label>
</template>

<style scoped>
.filter {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.icon {
  position: absolute;
  left: 0.55rem;
  color: var(--color-text-muted);
  pointer-events: none;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}

input {
  padding: 0.4rem 0.65rem 0.4rem 1.9rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: inherit;
  font: inherit;
  min-width: 16rem;
}

input:focus {
  outline: none;
  border-color: var(--color-accent);
  box-shadow: 0 0 0 2px var(--color-accent-subtle);
}
</style>
