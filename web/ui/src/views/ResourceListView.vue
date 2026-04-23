<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, type RouteLocationRaw } from 'vue-router';

import ResourceListTemplate from '@/components/resources/ResourceListTemplate.vue';
import DataTable, { type Column } from '@/components/ui/DataTable.vue';
import StatusPill from '@/components/ui/StatusPill.vue';
import FilterInput from '@/components/ui/FilterInput.vue';
import ActionMenu from '@/components/ui/ActionMenu.vue';
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
const deleting = ref(false);

const kind = computed(() => {
  const resourceId = String(route.params.resource ?? '');
  return resourceKindById(resourceId);
});

const columns = computed<Column<Row>[]>(() => {
  const base: Column<Row>[] = [
    { key: 'status', labelKey: 'columns.status', width: '9rem' },
    { key: 'name', labelKey: 'columns.name', get: (r) => r.name },
    { key: 'kind', labelKey: 'columns.kind', get: (r) => r.kind },
  ];
  base.push({ key: 'namespace', labelKey: 'columns.namespace', get: (r) => r.namespace || '—' });
  base.push({ key: 'created', labelKey: 'columns.created', get: (r) => r.created });
  return base;
});

const filteredRows = computed<readonly Row[]>(() => {
  const q = filter.value.trim().toLowerCase();
  const ns = ui.namespace;
  return rows.value.filter((r) => {
    if (ns && r.namespace !== ns) return false;
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
      // Kinds without a category in the aggregated endpoint (XRDs, ProviderConfigs)
      // are not listable via that endpoint yet. Empty for now — users can still
      // create via the form template.
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

watch(
  () => route.params.resource,
  () => {
    filter.value = '';
    void load();
  },
);
watch(
  () => ui.namespace,
  () => {
    // namespace filter is client-side; nothing to reload, computed handles it.
  },
);

onMounted(load);
</script>

<template>
  <ResourceListTemplate
    :title="title"
    :subtitle="subtitle"
    :loading="loading"
    :error="error"
    :count="filteredRows.length"
    :create-to="createTo"
    @refresh="load"
  >
    <template #toolbar>
      <FilterInput v-model="filter" />
    </template>

    <DataTable :rows="filteredRows" :columns="columns" :row-to="rowTo" empty-key="resource.list.empty">
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
  </ResourceListTemplate>
</template>
