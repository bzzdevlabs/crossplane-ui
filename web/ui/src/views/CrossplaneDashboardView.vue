<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Plug, Plus, RotateCw } from 'lucide-vue-next';

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

const allResources = computed<readonly CrossplaneResource[]>(() =>
  groups.value.flatMap((g) => g.items),
);

const totals = computed(() => {
  let ready = 0;
  let degraded = 0;
  let pending = 0;
  for (const r of allResources.value) {
    const s = statusFromConditions(r);
    if (s === 'ready') ready += 1;
    else if (s === 'degraded' || s === 'errored') degraded += 1;
    else pending += 1;
  }
  return { total: allResources.value.length, ready, degraded, pending };
});

const recent = computed<readonly CrossplaneResource[]>(() =>
  [...allResources.value]
    .sort(
      (a, b) =>
        new Date(b.creationTimestamp).getTime() -
        new Date(a.creationTimestamp).getTime(),
    )
    .slice(0, 5),
);

const providers = computed<readonly CrossplaneResource[]>(
  () => groups.value.find((g) => g.category === 'provider')?.items ?? [],
);

function age(ts: string): string {
  const ms = Date.now() - new Date(ts).getTime();
  if (Number.isNaN(ms) || ms < 0) return '';
  const days = Math.floor(ms / 86_400_000);
  if (days >= 1) return `${days}d`;
  const hours = Math.floor(ms / 3_600_000);
  if (hours >= 1) return `${hours}h`;
  const minutes = Math.floor(ms / 60_000);
  return `${Math.max(minutes, 1)}m`;
}

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
  <section class="dashboard">
    <header class="page-header">
      <div>
        <h1>{{ t('products.crossplane.dashboard') }}</h1>
        <p class="muted">{{ t('home.crossplaneHint') }}</p>
      </div>
      <div class="actions">
        <button type="button" class="btn neutral" :disabled="loading" @click="load">
          <RotateCw :size="14" aria-hidden="true" />
          {{ t('common.refresh') }}
        </button>
        <RouterLink :to="{ name: 'resource-create' }" class="btn primary">
          <Plus :size="14" aria-hidden="true" />
          {{ t('common.create') }}
        </RouterLink>
      </div>
    </header>

    <div class="stats">
      <div class="stat">
        <div class="label">{{ t('dashboard.allResources') }}</div>
        <div class="value">{{ totals.total }}</div>
        <div class="hint">{{ t('dashboard.allResourcesHint') }}</div>
      </div>
      <div class="stat">
        <div class="label">{{ t('status.variants.ready') }}</div>
        <div class="value ready">{{ totals.ready }}</div>
        <div class="hint">{{ t('dashboard.readyHint') }}</div>
      </div>
      <div class="stat">
        <div class="label">{{ t('status.variants.pending') }}</div>
        <div class="value pending">{{ totals.pending }}</div>
        <div class="hint">{{ t('dashboard.pendingHint') }}</div>
      </div>
      <div class="stat">
        <div class="label">{{ t('dashboard.needsAttention') }}</div>
        <div class="value degraded">{{ totals.degraded }}</div>
        <div class="hint">{{ t('dashboard.needsAttentionHint') }}</div>
      </div>
    </div>

    <p v-if="loading" class="muted">{{ t('home.loading') }}</p>
    <p v-else-if="error" class="error">{{ t('common.error') }}: {{ error }}</p>

    <div v-else class="split">
      <section class="card">
        <header class="card-header">
          <h2>{{ t('dashboard.recent') }}</h2>
          <RouterLink
            class="link"
            :to="{ name: 'resource-list', params: { resource: 'managed' } }"
          >
            {{ t('dashboard.viewAll') }} →
          </RouterLink>
        </header>
        <table>
          <thead>
            <tr>
              <th class="col-status">{{ t('columns.status') }}</th>
              <th>{{ t('columns.name') }}</th>
              <th>{{ t('columns.kind') }}</th>
              <th>{{ t('columns.namespace') }}</th>
              <th class="col-age">{{ t('columns.age') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in recent" :key="resourceKey(r)">
              <td>
                <StatusPill :variant="statusFromConditions(r)" />
              </td>
              <td>
                <RouterLink class="link" :to="resourceRoute(r)">{{ r.name }}</RouterLink>
              </td>
              <td>{{ r.kind }}</td>
              <td class="muted-cell">{{ r.namespace ?? '—' }}</td>
              <td class="muted-cell">{{ age(r.creationTimestamp) }}</td>
            </tr>
            <tr v-if="recent.length === 0">
              <td class="empty" colspan="5">{{ t('home.empty') }}</td>
            </tr>
          </tbody>
        </table>
      </section>

      <section class="card">
        <header class="card-header">
          <h2>{{ t('dashboard.providers') }}</h2>
        </header>
        <ul class="providers">
          <li v-for="p in providers" :key="resourceKey(p)" class="provider">
            <Plug :size="16" aria-hidden="true" />
            <div class="provider-text">
              <div class="provider-name">{{ p.name }}</div>
              <div class="provider-version">{{ p.apiVersion }}</div>
            </div>
            <StatusPill :variant="statusFromConditions(p)" />
          </li>
          <li v-if="providers.length === 0" class="provider empty">
            {{ t('home.emptyCategory') }}
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
  gap: 1.25rem;
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
  font-weight: 600;
}

.muted {
  margin: 0.25rem 0 0;
  color: var(--color-text-muted);
  font-size: 0.9rem;
}

.error {
  color: var(--color-danger);
}

.actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.4rem 0.9rem;
  border-radius: 6px;
  font: inherit;
  font-size: 0.875rem;
  cursor: pointer;
  text-decoration: none;
}

.btn.neutral {
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: inherit;
}

.btn.primary {
  border: 1px solid var(--color-accent);
  background: var(--color-accent);
  color: var(--color-on-accent);
}

.btn[disabled] {
  opacity: 0.6;
  cursor: progress;
}

.stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.75rem;
}

@media (max-width: 60rem) {
  .stats {
    grid-template-columns: repeat(2, 1fr);
  }
}

.stat {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 0.9rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.stat .label {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-weight: 600;
  color: var(--color-text-muted);
}

.stat .value {
  font-size: 1.6rem;
  font-weight: 600;
  color: var(--color-text);
}

.stat .value.ready {
  color: #1f7a3a;
}

.stat .value.pending {
  color: #8a6100;
}

.stat .value.degraded {
  color: #a3341f;
}

.stat .hint {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.split {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 1rem;
}

@media (max-width: 60rem) {
  .split {
    grid-template-columns: 1fr;
  }
}

.card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 0.9rem;
  border-bottom: 1px solid var(--color-border);
}

.card-header h2 {
  margin: 0;
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--color-text-muted);
}

.link {
  color: var(--color-accent);
  text-decoration: none;
  font-size: 0.8rem;
}

.link:hover {
  text-decoration: underline;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th {
  text-align: left;
  padding: 0.55rem 0.9rem;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  background: var(--color-surface-alt);
  border-bottom: 1px solid var(--color-border);
}

td {
  padding: 0.55rem 0.9rem;
  border-bottom: 1px solid var(--color-border);
  font-size: 0.875rem;
  vertical-align: middle;
}

tr:last-child td {
  border-bottom: 0;
}

.col-status {
  width: 8.5rem;
}

.col-age {
  width: 5rem;
}

.muted-cell {
  color: var(--color-text-muted);
}

.empty {
  text-align: center;
  padding: 2rem;
  color: var(--color-text-muted);
}

.providers {
  list-style: none;
  padding: 0;
  margin: 0;
}

.provider {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.9rem;
  border-bottom: 1px solid var(--color-border);
}

.provider:last-child {
  border-bottom: 0;
}

.provider-text {
  flex: 1;
  min-width: 0;
}

.provider-name {
  font-size: 0.875rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.provider-version {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  font-family: var(--font-mono);
}

.provider.empty {
  justify-content: center;
  font-size: 0.85rem;
  color: var(--color-text-muted);
}
</style>
