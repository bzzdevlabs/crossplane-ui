<script setup lang="ts">
export interface TabDescriptor {
  readonly id: string;
  readonly label: string;
  readonly disabled?: boolean;
}

defineProps<{
  tabs: readonly TabDescriptor[];
  modelValue: string;
}>();

defineEmits<{
  (e: 'update:modelValue', value: string): void;
}>();
</script>

<template>
  <div class="tabs" role="tablist">
    <button
      v-for="tab in tabs"
      :key="tab.id"
      type="button"
      role="tab"
      :aria-selected="modelValue === tab.id"
      :disabled="tab.disabled"
      :class="{ active: modelValue === tab.id }"
      @click="$emit('update:modelValue', tab.id)"
    >
      {{ tab.label }}
    </button>
  </div>
</template>

<style scoped>
.tabs {
  display: flex;
  gap: 0.25rem;
  border-bottom: 1px solid var(--color-border);
}

button {
  appearance: none;
  background: transparent;
  border: 0;
  border-bottom: 2px solid transparent;
  color: var(--color-text-muted);
  padding: 0.55rem 1rem;
  font: inherit;
  font-size: 0.9rem;
  cursor: pointer;
  margin-bottom: -1px;
}

button:hover:not([disabled]) {
  color: var(--color-text);
}

button.active {
  color: var(--color-accent);
  border-bottom-color: var(--color-accent);
  font-weight: 500;
}

button[disabled] {
  opacity: 0.45;
  cursor: not-allowed;
}
</style>
