<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { ChevronDown } from 'lucide-vue-next';

import { useUiStore } from '@/stores/ui';

const ui = useUiStore();
const { t } = useI18n();

const open = ref(false);
const root = ref<HTMLElement | null>(null);
const query = ref('');

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  if (!q) return ui.namespaces;
  return ui.namespaces.filter((n) => n.name.toLowerCase().includes(q));
});

const currentLabel = computed(() =>
  ui.namespace === null ? t('namespace.all') : ui.namespace,
);

function toggle(): void {
  open.value = !open.value;
  if (open.value) void ui.loadNamespaces();
}

function close(): void {
  open.value = false;
  query.value = '';
}

function pick(ns: string | null): void {
  ui.setNamespace(ns);
  close();
}

function handleDocClick(e: MouseEvent): void {
  if (!root.value) return;
  if (!root.value.contains(e.target as Node)) close();
}

onMounted(() => document.addEventListener('click', handleDocClick));
onBeforeUnmount(() => document.removeEventListener('click', handleDocClick));

watch(
  () => ui.productSwitcherOpen,
  (v) => {
    if (v) close();
  },
);
</script>

<template>
  <div ref="root" class="picker">
    <button
      type="button"
      class="trigger"
      :aria-expanded="open"
      :aria-label="t('namespace.picker')"
      @click="toggle"
    >
      <span class="label">{{ t('namespace.picker') }}</span>
      <span class="value">{{ currentLabel }}</span>
      <ChevronDown class="caret" :size="14" aria-hidden="true" />
    </button>

    <div v-if="open" class="popover" role="listbox">
      <input
        v-model="query"
        type="search"
        class="search"
        :placeholder="t('common.filter')"
      />
      <ul>
        <li>
          <button
            type="button"
            role="option"
            :aria-selected="ui.namespace === null"
            :class="{ selected: ui.namespace === null }"
            @click="pick(null)"
          >
            {{ t('namespace.all') }}
          </button>
        </li>
        <li v-if="ui.namespacesLoading" class="muted">{{ t('namespace.loading') }}</li>
        <li v-else-if="ui.namespacesError" class="error">
          {{ t('namespace.errorLoading') }}
        </li>
        <li v-for="ns in filtered" v-else :key="ns.name">
          <button
            type="button"
            role="option"
            :aria-selected="ui.namespace === ns.name"
            :class="{ selected: ui.namespace === ns.name }"
            @click="pick(ns.name)"
          >
            {{ ns.name }}
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.picker {
  position: relative;
}

.trigger {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.35rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: inherit;
  font: inherit;
  cursor: pointer;
  font-size: 0.85rem;
}

.label {
  color: var(--color-text-muted);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.value {
  font-weight: 500;
  max-width: 14rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.caret {
  color: var(--color-text-muted);
}

.popover {
  position: absolute;
  right: 0;
  top: calc(100% + 0.25rem);
  z-index: 15;
  width: 18rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  box-shadow: 0 6px 18px rgb(0 0 0 / 12%);
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.search {
  width: 100%;
  padding: 0.35rem 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  font: inherit;
  color: inherit;
}

ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  max-height: 20rem;
  overflow-y: auto;
}

ul button {
  width: 100%;
  text-align: left;
  background: transparent;
  border: 0;
  color: inherit;
  font: inherit;
  padding: 0.35rem 0.5rem;
  border-radius: 4px;
  cursor: pointer;
}

ul button:hover,
ul button.selected {
  background: var(--color-accent-subtle);
  color: var(--color-accent);
}

.muted,
.error {
  color: var(--color-text-muted);
  padding: 0.35rem 0.5rem;
  font-size: 0.85rem;
}

.error {
  color: var(--color-danger);
}
</style>
