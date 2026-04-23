<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Plus, RotateCw } from 'lucide-vue-next';

import StatusPill from '@/components/ui/StatusPill.vue';
import { statusFromConditions } from '@/resources/registry';
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

const totals = computed(() => {
  let ready = 0;
  let degraded = 0;
  let pending = 0;
  let total = 0;
  for (const g of groups.value) {
    for (const r of g.items) {
      total += 1;
      const s = statusFromConditions(r);
      if (s === 'ready') ready += 1;
      else if (s === 'degraded' || s === 'errored') degraded += 1;
      else pending += 1;
    }
  }
  return { ready, degraded, pending, total };
});

const sortedGroups = computed(() => {
  const idx = (c: string): number => {
    const i = (CATEGORY_ORDER as readonly string[]).indexOf(c);
    return i < 0 ? CATEGORY_ORDER.length : i;
  };
  return [...groups.value].sort((a, b) => idx(a.category) - idx(b.category));
});

function resourceRoute(r: CrossplaneResource) {
  const parts = r.apiVersion.split('/');
  const group = parts.length > 1 ? parts[0] ?? '' : '';
  const version = parts.length > 1 ? parts[1] ?? '' : parts[0] ?? '';
  const query: Record<string, string> = {};
  if (group) query.group = group;
  if (version) query.version = version;
  if (r.namespace) query.namespace = r.namespace;
  return {
    name: 'resource-detail',
    params: { resource: r.resource, name: r.name },
    query,
  } as const;
}

function listRouteForCategory(category: string) {
  const kindId = (
    {
      composition: 'compositions',
      composite: 'composites',
      managed: 'managed',
      provider: 'providers',
      function: 'functions',
    } as Record<string, string>
  )[category];
  return { name: 'resource-list', params: { resource: kindId ?? category } };
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

function kindLabelForCategory(category: string): string {
  const map: Record<string, string> = {
    composition: 'kinds.composition.plural',
    composite: 'kinds.composite.plural',
    managed: 'kinds.managed.plural',
    provider: 'kinds.provider.plural',
    function: 'kinds.function.plural',
  };
  return t(map[category] ?? category);
}

function resourceKey(r: CrossplaneResource): string {
  return `${r.apiVersion}|${r.kind}|${r.namespace ?? ''}|${r.name}`;
}

onMounted(load);
</script>

<template>
  <section class="dashboard">
    <header class="page-header">
      <div>
        <h1>{{ t('products.crossplane.dashboard') }}</h1>
        <p class="muted">{{ t('home.crossplaneHint') }}</p>
      </div>
      <div class="actions">
        <button type="button" class="refresh" :disabled="loading" @click="load">
          <RotateCw :size="14" aria-hidden="true" />
          {{ t('common.refresh') }}
        </button>
        <RouterLink :to="{ name: 'resource-create' }" class="create">
          <Plus :size="14" aria-hidden="true" />
          {{ t('resource.create') }}
        </RouterLink>
      </div>
    </header>

    <div v-if="!loading && !error" class="stats">
      <div class="stat">
        <div class="stat-value">{{ totals.total }}</div>
        <div class="stat-label">{{ t('resource.create') }}</div>
      </div>
      <div class="stat ready">
        <div class="stat-value">{{ totals.ready }}</div>
        <div class="stat-label">{{ t('status.variants.ready') }}</div>
      </div>
      <div class="stat degraded">
        <div class="stat-value">{{ totals.degraded }}</div>
        <div class="stat-label">{{ t('status.variants.degraded') }}</div>
      </div>
      <div class="stat pending">
        <div class="stat-value">{{ totals.pending }}</div>
        <div class="stat-label">{{ t('status.variants.pending') }}</div>
      </div>
    </div>

    <p v-if="loading" class="muted">{{ t('home.loading') }}</p>
    <p v-else-if="error" class="error">{{ t('common.error') }}: {{ error }}</p>

    <div v-else class="groups">
      <section v-for="g in sortedGroups" :key="g.category" class="group">
        <header class="group-header">
          <RouterLink :to="listRouteForCategory(g.category)">
            <h2>{{ kindLabelForCategory(g.category) }}</h2>
          </RouterLink>
          <span class="count">{{ g.items.length }}</span>
        </header>

        <p v-if="g.error" class="group-error">{{ g.error }}</p>

        <p v-if="!g.error && g.items.length === 0" class="muted small">
          {{ t('home.emptyCategory') }}
        </p>

        <ul v-else-if="g.items.length > 0" class="tiles">
          <li v-for="r in g.items" :key="resourceKey(r)">
            <RouterLink class="tile" :to="resourceRoute(r)">
              <div class="tile-head">
                <div class="tile-title">{{ r.name }}</div>
                <div class="tile-kind">{{ r.kind }}</div>
              </div>
              <div class="tile-meta">
                <StatusPill :variant="statusFromConditions(r)" />
                <time :datetime="r.creationTimestamp">{{
                  new Date(r.creationTimestamp).toLocaleDateString()
                }}</time>
              </div>
              <div v-if="r.namespace" class="tile-ns">{{ r.namespace }}</div>
            </RouterLink>
          </li>
        </ul>
      </section>
    </div>
  </section>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
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
  color: inherit;
}

.muted {
  margin: 0.25rem 0 0;
  color: var(--color-text-muted);
}

.small {
  font-size: 0.85rem;
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

.refresh[disabled] {
  opacity: 0.6;
  cursor: progress;
}

.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 0.75rem;
}

.stat {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 0.9rem 1rem;
}

.stat-value {
  font-size: 1.6rem;
  font-weight: 600;
}

.stat-label {
  color: var(--color-text-muted);
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.stat.ready {
  border-color: #1f7a3a;
}

.stat.degraded {
  border-color: #a3341f;
}

.stat.pending {
  border-color: #c9a227;
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

.group-header a {
  color: inherit;
  text-decoration: none;
}

.group-header a:hover h2 {
  color: var(--color-accent);
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
  color: inherit;
  text-decoration: none;
  transition: border-color 0.1s ease;
}

.tile:hover {
  border-color: var(--color-accent);
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

.tile-ns {
  color: var(--color-text-muted);
  font-size: 0.75rem;
}
</style>
