<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import FormShell from '@/components/forms/FormShell.vue';
import type { Obj } from '@/components/forms/path';
import { schemaForObject } from '@/components/forms/schemas';
import ResourceDetailTemplate from '@/components/resources/ResourceDetailTemplate.vue';
import Tabs from '@/components/ui/Tabs.vue';
import { resourceKindById, statusFromConditions } from '@/resources/registry';
import type { StatusVariant } from '@/resources/registry';
import {
  applyResource,
  deleteResource,
  getResource,
  type CrossplaneResource,
  type ResourceRef,
} from '@/services/api';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const loading = ref(false);
const saving = ref(false);
const errorMsg = ref<string | null>(null);
const notice = ref<string | null>(null);
const original = ref<Obj>({});
const draft = ref<Obj>({});
const activeTab = ref<'details' | 'yaml'>('details');

const ref_ = computed<ResourceRef>(() => {
  const params = route.params as Record<string, string | string[] | undefined>;
  const query = route.query as Record<string, string | string[] | undefined>;
  const resourceSlug = String(params.resource ?? '');
  const kind = resourceKindById(resourceSlug);
  const group = kind?.gvr?.group ?? (Array.isArray(query.group) ? query.group[0] : query.group ?? '');
  const version =
    kind?.gvr?.version ?? (Array.isArray(query.version) ? query.version[0] : query.version ?? '');
  const resource = kind?.gvr?.resource ?? resourceSlug;
  const namespaceRaw = Array.isArray(query.namespace) ? query.namespace[0] : query.namespace;
  return {
    group: group ?? '',
    version: version ?? '',
    resource,
    name: String(params.name ?? ''),
    namespace: namespaceRaw || undefined,
  };
});

const schema = computed(() => schemaForObject(draft.value));
const dirty = computed(() => JSON.stringify(draft.value) !== JSON.stringify(original.value));

const title = computed(() => ref_.value.name);
const kindLabel = computed(() => {
  const obj = draft.value;
  if (typeof obj.kind === 'string') return obj.kind;
  return ref_.value.resource;
});

const metaParts = computed(() => {
  const parts: string[] = [];
  if (ref_.value.group) parts.push(`${ref_.value.group}/${ref_.value.version}`);
  else if (ref_.value.version) parts.push(ref_.value.version);
  parts.push(ref_.value.resource);
  if (ref_.value.namespace) parts.push(ref_.value.namespace);
  return parts;
});

const status = computed<StatusVariant | undefined>(() => {
  const obj: Record<string, unknown> = draft.value;
  const statusObj = obj.status ?? original.value.status;
  const conds =
    statusObj && typeof statusObj === 'object'
      ? (statusObj as { conditions?: unknown }).conditions
      : undefined;
  if (!Array.isArray(conds)) return undefined;
  const apiVersion = typeof obj.apiVersion === 'string' ? obj.apiVersion : '';
  const kind = typeof obj.kind === 'string' ? obj.kind : '';
  const mock: CrossplaneResource = {
    apiVersion,
    kind,
    resource: ref_.value.resource,
    name: ref_.value.name,
    ready: 'Unknown',
    synced: 'Unknown',
    creationTimestamp: '',
  };
  let ready = 'Unknown';
  let synced = 'Unknown';
  for (const c of conds) {
    if (!c || typeof c !== 'object') continue;
    const rec = c as Record<string, unknown>;
    const type = typeof rec.type === 'string' ? rec.type : '';
    const st = typeof rec.status === 'string' ? rec.status : 'Unknown';
    if (type === 'Ready') ready = st;
    if (type === 'Synced') synced = st;
  }
  return statusFromConditions({ ...mock, ready, synced });
});

function stripServerFields(obj: Obj): Obj {
  const clone = structuredClone(obj);
  const meta = (clone.metadata ?? {}) as Obj;
  delete meta.managedFields;
  delete meta.resourceVersion;
  delete meta.uid;
  delete meta.generation;
  delete meta.selfLink;
  delete meta.creationTimestamp;
  delete clone.status;
  return clone;
}

async function load(): Promise<void> {
  loading.value = true;
  errorMsg.value = null;
  notice.value = null;
  try {
    const obj = await getResource<Obj>(ref_.value);
    const editable = stripServerFields(obj);
    original.value = editable;
    draft.value = structuredClone(editable);
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

async function save(): Promise<void> {
  if (saving.value) return;
  saving.value = true;
  errorMsg.value = null;
  notice.value = null;
  try {
    const applied = await applyResource<Obj>(ref_.value, draft.value);
    const editable = stripServerFields(applied);
    original.value = editable;
    draft.value = structuredClone(editable);
    notice.value = t('resource.saved');
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err);
  } finally {
    saving.value = false;
  }
}

async function remove(): Promise<void> {
  if (!window.confirm(t('resource.confirmDelete', { name: ref_.value.name }))) return;
  saving.value = true;
  errorMsg.value = null;
  try {
    await deleteResource(ref_.value);
    await router.replace({ name: 'crossplane-dashboard' });
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err);
    saving.value = false;
  }
}

const tabs = computed(() => [
  { id: 'details', label: t('resource.tabs.details') },
  { id: 'yaml', label: t('resource.tabs.yaml') },
]);

const breadcrumbs = computed(() => [
  { label: t('products.crossplane.label'), to: { name: 'crossplane-dashboard' } },
  { label: ref_.value.resource },
  { label: ref_.value.name },
]);

watch(ref_, load, { immediate: false });
onMounted(load);
</script>

<template>
  <ResourceDetailTemplate
    :title="title"
    :kind="kindLabel"
    :meta-parts="metaParts"
    :status="status"
    :breadcrumbs="breadcrumbs"
    :saving="saving"
    :can-apply="dirty"
    @refresh="load"
    @delete="remove"
    @apply="save"
  >
    <p v-if="errorMsg" class="error">{{ errorMsg }}</p>
    <p v-if="notice" class="notice">{{ notice }}</p>

    <p v-if="loading" class="muted">{{ t('home.loading') }}</p>
    <template v-else>
      <Tabs v-model="activeTab" :tabs="tabs" />
      <div class="tab-body">
        <FormShell
          v-if="activeTab === 'details'"
          v-model="draft"
          :form-component="schema?.component ?? null"
        />
        <FormShell v-else v-model="draft" :form-component="null" />
      </div>
    </template>
  </ResourceDetailTemplate>
</template>

<style scoped>
.error {
  color: var(--color-danger);
  margin: 0;
  white-space: pre-wrap;
}

.notice {
  color: #1f7a3a;
  margin: 0;
}

.muted {
  margin: 0;
  color: var(--color-text-muted);
}

.tab-body {
  margin-top: 0.75rem;
}
</style>
