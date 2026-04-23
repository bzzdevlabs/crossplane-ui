<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { MoreVertical } from 'lucide-vue-next';

export interface ActionItem {
  readonly id: string;
  readonly label: string;
  readonly danger?: boolean;
  readonly disabled?: boolean;
}

defineProps<{
  items: readonly ActionItem[];
  label?: string;
}>();

const emit = defineEmits<{
  (e: 'select', id: string): void;
}>();

const open = ref(false);
const root = ref<HTMLElement | null>(null);

function toggle(): void {
  open.value = !open.value;
}

function close(): void {
  open.value = false;
}

function onPick(id: string): void {
  emit('select', id);
  close();
}

function handleDocClick(e: MouseEvent): void {
  if (!root.value) return;
  if (!root.value.contains(e.target as Node)) close();
}

onMounted(() => document.addEventListener('click', handleDocClick));
onBeforeUnmount(() => document.removeEventListener('click', handleDocClick));
</script>

<template>
  <div ref="root" class="menu">
    <button
      type="button"
      class="trigger"
      :aria-label="label ?? 'Actions'"
      :aria-expanded="open"
      @click="toggle"
    >
      <MoreVertical :size="16" aria-hidden="true" />
    </button>
    <ul v-if="open" class="items" role="menu">
      <li v-for="item in items" :key="item.id">
        <button
          type="button"
          role="menuitem"
          :disabled="item.disabled"
          :class="{ danger: item.danger }"
          @click="onPick(item.id)"
        >
          {{ item.label }}
        </button>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.menu {
  position: relative;
  display: inline-flex;
}

.trigger {
  width: 2rem;
  height: 2rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: inherit;
  border-radius: 6px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.trigger:hover {
  background: var(--color-surface-alt);
}

.items {
  position: absolute;
  right: 0;
  top: calc(100% + 0.25rem);
  z-index: 10;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 0.25rem;
  list-style: none;
  margin: 0;
  min-width: 10rem;
  box-shadow: 0 4px 12px rgb(0 0 0 / 12%);
}

.items button {
  width: 100%;
  text-align: left;
  background: transparent;
  border: 0;
  color: inherit;
  font: inherit;
  padding: 0.4rem 0.6rem;
  border-radius: 4px;
  cursor: pointer;
}

.items button:hover:not([disabled]) {
  background: var(--color-accent-subtle);
}

.items button.danger {
  color: var(--color-danger);
}

.items button[disabled] {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>
