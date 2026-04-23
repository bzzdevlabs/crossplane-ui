<script setup lang="ts">
defineProps<{
  label: string;
  hint?: string;
  required?: boolean;
}>();
</script>

<template>
  <label class="field">
    <span class="label">
      {{ label }}
      <span v-if="required" class="req" aria-hidden="true">*</span>
    </span>
    <slot />
    <span v-if="hint" class="hint">{{ hint }}</span>
  </label>
</template>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.85rem;
}

.label {
  color: var(--color-text-muted);
  font-weight: 500;
}

.req {
  color: var(--color-danger);
  margin-left: 0.15rem;
}

.hint {
  color: var(--color-text-muted);
  font-size: 0.75rem;
}

.field :deep(input),
.field :deep(select),
.field :deep(textarea) {
  padding: 0.4rem 0.55rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: inherit;
  font: inherit;
  font-size: 0.9rem;
}

.field :deep(textarea) {
  min-height: 4rem;
  resize: vertical;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.8rem;
}

.field :deep(input:focus),
.field :deep(select:focus),
.field :deep(textarea:focus) {
  outline: 2px solid var(--color-accent);
  outline-offset: -1px;
}
</style>
