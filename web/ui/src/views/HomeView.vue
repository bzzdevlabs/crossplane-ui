<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import {
  listCrossplaneResources,
  type CrossplaneGroup,
  type CrossplaneResource,
} from '@/services/api';

const { t } = useI18n();

const loading = ref(false);
const error = ref<string | null>(null);
const groups = ref<readonly CrossplaneGroup[]>([]);

const CATEGORY_ORDER = ['composition', 'composite', 'managed', 'provider', 'function'] as const;

const sortedGroups = computed(() => {
  const idx = (c: string) => {
    const i = (CATEGORY_ORDER as readonly string[]).indexOf(c);
    return i < 0 ? CATEGORY_ORDER.length : i;
  };
  return [...groups.value].sort((a, b) => idx(a.category) - idx(b.category));
});

function badgeClass(status: string): string {
  switch (status) {
    case 'True':
      return 'badge badge-ok';
    case 'False':
      return 'badge badge-bad';
    default:
      return 'badge badge-unknown';
  }
}

function badgeLabel(status: string): string {
  switch (status) {
    case 'True':
      return t('status.ready');
    case 'False':
      return t('status.notReady');
    default:
      return t('status.unknown');
  }
}

function resourceKey(r: CrossplaneResource): string {
  return `${r.apiVersion}|${r.kind}|${r.namespace ?? ''}|${r.name}`;
}

async function load(): Promise<void> {
  loading.value = true;
  error.value = null;
  try {
    const response = await listCrossplaneResources();
    groups.value = response.groups;
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <section class="home">
    <header class="page-header">
      <div>
        <h1>{{ t('home.heading') }}</h1>
        <p class="muted">{{ t('home.crossplaneHint') }}</p>
      </div>
      <button type="button" class="refresh" :disabled="loading" @click="load">
        {{ t('home.refresh') }}
      </button>
    </header>

    <p v-if="loading" class="muted">{{ t('home.loading') }}</p>
    <p v-else-if="error" class="error">{{ t('common.error') }}: {{ error }}</p>

    <div v-else class="groups">
      <section v-for="g in sortedGroups" :key="g.category" class="group">
        <header class="group-header">
          <h2>{{ t(`home.categories.${g.category}`) }}</h2>
          <span class="count">{{ g.items.length }}</span>
        </header>

        <p v-if="g.error" class="group-error">{{ g.error }}</p>

        <p v-if="!g.error && g.items.length === 0" class="muted small">
          {{ t('home.emptyCategory') }}
        </p>

        <ul v-else-if="g.items.length > 0" class="tiles">
          <li v-for="r in g.items" :key="resourceKey(r)" class="tile">
            <div class="tile-head">
              <div class="tile-title">{{ r.name }}</div>
              <div class="tile-kind">{{ r.kind }}</div>
            </div>
            <div class="tile-meta">
              <div class="badges">
                <span :class="badgeClass(r.ready)" :title="`Ready=${r.ready}`">
                  {{ badgeLabel(r.ready) }}
                </span>
                <span
                  :class="badgeClass(r.synced)"
                  :title="`Synced=${r.synced}`"
                >
                  {{ r.synced === 'True' ? t('status.synced') : t('status.outOfSync') }}
                </span>
              </div>
              <time :datetime="r.creationTimestamp">{{
                new Date(r.creationTimestamp).toLocaleDateString()
              }}</time>
            </div>
            <div v-if="r.namespace" class="tile-ns">{{ r.namespace }}</div>
          </li>
        </ul>
      </section>
    </div>
  </section>
</template>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
}

h1 {
  margin: 0;
  font-size: 1.5rem;
}

h2 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.muted {
  margin: 0;
  color: var(--color-text-muted);
}

.small {
  font-size: 0.85rem;
}

.error {
  color: var(--color-danger);
}

.refresh {
  padding: 0.4rem 0.9rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.refresh[disabled] {
  opacity: 0.6;
  cursor: progress;
}

.groups {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.group-header {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.count {
  color: var(--color-text-muted);
  font-size: 0.85rem;
}

.group-error {
  color: var(--color-danger);
  font-size: 0.85rem;
  margin: 0;
}

.tiles {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 0.75rem;
}

.tile {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 0.9rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.tile-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 0.5rem;
}

.tile-title {
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tile-kind {
  color: var(--color-text-muted);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.tile-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--color-text-muted);
  font-size: 0.8rem;
}

.badges {
  display: flex;
  gap: 0.35rem;
  flex-wrap: wrap;
}

.badge {
  padding: 0.1rem 0.5rem;
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 500;
}

.badge-ok {
  background: #e6f6ed;
  color: #1f7a3a;
}

.badge-bad {
  background: #fdebe7;
  color: #a3341f;
}

.badge-unknown {
  background: var(--color-surface-alt);
  color: var(--color-text-muted);
}

.tile-ns {
  color: var(--color-text-muted);
  font-size: 0.75rem;
}
</style>
