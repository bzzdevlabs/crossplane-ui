<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, type RouteLocationRaw } from 'vue-router';
import { Plus, RotateCw } from 'lucide-vue-next';

import ActionMenu from '@/components/ui/ActionMenu.vue';
import BreadcrumbBar from '@/components/ui/BreadcrumbBar.vue';
import DataTable, { type Column } from '@/components/ui/DataTable.vue';
import FilterInput from '@/components/ui/FilterInput.vue';
import StatusFilterGroup, {
  type StatusFilterValue,
} from '@/components/ui/StatusFilterGroup.vue';
import StatusPill from '@/components/ui/StatusPill.vue';
import { resourceKindById, statusFromConditions } from '@/resources/registry';
import { useUiStore } from '@/stores/ui';
import {
  deleteResource,
  listCrossplaneResources,
  type CrossplaneResource,
} from '@/services/api';

interface Row {
  readonly id: string;
  readonly name: string;
  readonly kind: string;
  readonly namespace: string;
  readonly created: string;
  readonly resource: CrossplaneResource;
}

const { t } = useI18n();
const route = useRoute();
const ui = useUiStore();

const rows = ref<readonly Row[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const filter = ref('');
const statusFilter = ref<StatusFilterValue>('all');
const deleting = ref(false);

const kind = computed(() => {
  const resourceId = String(route.params.resource ?? '');
  return resourceKindById(resourceId);
});

const columns = computed<Column<Row>[]>(() => [
  { key: 'status', labelKey: 'columns.status', width: '8.5rem' },
  { key: 'name', labelKey: 'columns.name', get: (r) => r.name },
  { key: 'kind', labelKey: 'columns.kind', get: (r) => r.kind },
  { key: 'namespace', labelKey: 'columns.namespace', get: (r) => r.namespace || '—' },
  { key: 'created', labelKey: 'columns.created', get: (r) => r.created },
]);

const filteredRows = computed<readonly Row[]>(() => {
  const q = filter.value.trim().toLowerCase();
  const ns = ui.namespace;
  const sf = statusFilter.value;
  return rows.value.filter((r) => {
    if (ns && r.namespace !== ns) return false;
    if (sf !== 'all' && statusFromConditions(r.resource) !== sf) return false;
    if (q && !r.name.toLowerCase().includes(q) && !r.kind.toLowerCase().includes(q)) {
      return false;
    }
    return true;
  });
});

function rowTo(row: Row): RouteLocationRaw {
  return {
    name: 'resource-detail',
    params: { resource: row.resource.resource, name: row.resource.name },
    query: row.resource.namespace ? { namespace: row.resource.namespace } : undefined,
  };
}

async function load(): Promise<void> {
  const k = kind.value;
  if (!k) {
    rows.value = [];
    error.value = `Unknown kind: ${String(route.params.resource)}`;
    return;
  }
  loading.value = true;
  error.value = null;
  try {
    if (k.category) {
      const res = await listCrossplaneResources();
      const group = res.groups.find((g) => g.category === k.category);
      if (!group) {
        rows.value = [];
        return;
      }
      if (group.error) throw new Error(group.error);
      rows.value = group.items.map((item) => ({
        id: `${item.resource}|${item.namespace ?? ''}|${item.name}`,
        name: item.name,
        kind: item.kind,
        namespace: item.namespace ?? '',
        created: new Date(item.creationTimestamp).toLocaleString(),
        resource: item,
      }));
    } else {
      rows.value = [];
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

async function remove(row: Row): Promise<void> {
  if (!window.confirm(t('resource.confirmDelete', { name: row.name }))) return;
  const r = row.resource;
  const parts = r.apiVersion.split('/');
  const group = parts.length > 1 ? parts[0] ?? '' : '';
  const version = parts.length > 1 ? parts[1] ?? '' : parts[0] ?? '';
  deleting.value = true;
  try {
    await deleteResource({
      group,
      version,
      resource: r.resource,
      name: r.name,
      namespace: r.namespace,
    });
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    deleting.value = false;
  }
}

const title = computed(() => (kind.value ? t(kind.value.pluralLabelKey) : ''));
const subtitle = computed(() => (kind.value ? t(kind.value.labelKey) : ''));

const canCreate = computed(() => Boolean(kind.value?.form));
const createTo = computed<RouteLocationRaw | undefined>(() =>
  canCreate.value ? { name: 'resource-create', query: { kind: kind.value?.form?.id } } : undefined,
);

const breadcrumbs = computed(() => [
  { label: t('products.crossplane.label'), to: { name: 'crossplane-dashboard' } as RouteLocationRaw },
  { label: title.value },
]);

watch(
  () => route.params.resource,
  () => {
    filter.value = '';
    statusFilter.value = 'all';
    void load();
  },
);

onMounted(load);
</script>

<template>
  <section class="list-view">
    <BreadcrumbBar :items="breadcrumbs" />

    <header class="page-header">
      <div>
        <h1>{{ title }}</h1>
        <p v-if="subtitle" class="muted">{{ subtitle }}</p>
      </div>
      <div class="actions">
        <button type="button" class="btn neutral" :disabled="loading" @click="load">
          <RotateCw :size="14" aria-hidden="true" />
          {{ t('common.refresh') }}
        </button>
        <RouterLink v-if="createTo" :to="createTo" class="btn primary">
          <Plus :size="14" aria-hidden="true" />
          {{ t('resource.list.create') }}
        </RouterLink>
      </div>
    </header>

    <div class="filters">
      <FilterInput v-model="filter" />
      <StatusFilterGroup v-model="statusFilter" />
      <span class="count">{{ filteredRows.length }} of {{ rows.length }}</span>
    </div>

    <p v-if="loading" class="muted">{{ t('home.loading') }}</p>
    <p v-else-if="error" class="error">{{ t('common.error') }}: {{ error }}</p>

    <DataTable
      v-else
      :rows="filteredRows"
      :columns="columns"
      :row-to="rowTo"
      empty-key="resource.list.empty"
    >
      <template #cell-status="{ row }">
        <StatusPill :variant="statusFromConditions(row.resource)" />
      </template>
      <template #row-actions="{ row }">
        <ActionMenu
          :items="[{ id: 'delete', label: t('common.delete'), danger: true, disabled: deleting }]"
          @select="(id) => id === 'delete' && remove(row)"
        />
      </template>
    </DataTable>
  </section>
</template>

<style scoped>
.list-view {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 1rem;
  flex-wrap: wrap;
  margin-top: 0.5rem;
}

h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.muted {
  color: var(--color-text-muted);
  margin: 0.25rem 0 0;
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

.filters {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.count {
  margin-left: auto;
  font-size: 0.8rem;
  color: var(--color-text-muted);
}
</style>
