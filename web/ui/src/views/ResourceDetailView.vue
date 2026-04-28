<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import FormShell from '@/components/forms/FormShell.vue';
import type { Obj } from '@/components/forms/path';
import { schemaForObject } from '@/components/forms/schemas';
import ResourceDetailTemplate from '@/components/resources/ResourceDetailTemplate.vue';
import StatusPill from '@/components/ui/StatusPill.vue';
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
type TabId = 'details' | 'yaml' | 'events' | 'conditions';
const activeTab = ref<TabId>('details');

interface Condition {
  readonly type: string;
  readonly status: string;
  readonly reason?: string;
  readonly lastTransitionTime?: string;
}

function age(ts: string | undefined): string {
  if (!ts) return '';
  const ms = Date.now() - new Date(ts).getTime();
  if (Number.isNaN(ms) || ms < 0) return '';
  const days = Math.floor(ms / 86_400_000);
  if (days >= 1) return `${days}d`;
  const hours = Math.floor(ms / 3_600_000);
  if (hours >= 1) return `${hours}h`;
  return `${Math.max(Math.floor(ms / 60_000), 1)}m`;
}

const conditions = computed<readonly Condition[]>(() => {
  const obj = original.value;
  const statusObj = obj.status;
  const list =
    statusObj && typeof statusObj === 'object'
      ? (statusObj as { conditions?: unknown }).conditions
      : undefined;
  if (!Array.isArray(list)) return [];
  return list
    .filter((c): c is Record<string, unknown> => !!c && typeof c === 'object')
    .map((c) => ({
      type: typeof c.type === 'string' ? c.type : '',
      status: typeof c.status === 'string' ? c.status : 'Unknown',
      reason: typeof c.reason === 'string' ? c.reason : undefined,
      lastTransitionTime:
        typeof c.lastTransitionTime === 'string' ? c.lastTransitionTime : undefined,
    }));
});

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
  { id: 'events', label: t('resource.tabs.events') },
  { id: 'conditions', label: t('resource.tabs.conditions') },
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
        <FormShell v-else-if="activeTab === 'yaml'" v-model="draft" :form-component="null" />
        <section v-else-if="activeTab === 'conditions'" class="card">
          <div
            v-for="(c, i) in conditions"
            :key="`${c.type}-${i}`"
            class="cond-row"
          >
            <span class="cond-type">{{ c.type }}</span>
            <StatusPill
              :variant="c.status === 'True' ? 'ready' : 'degraded'"
              :label="c.status"
            />
            <span v-if="c.reason" class="cond-reason">{{ c.reason }}</span>
            <span class="cond-age">{{ age(c.lastTransitionTime) }}</span>
          </div>
          <p v-if="conditions.length === 0" class="empty">
            {{ t('resource.detail.noConditions') }}
          </p>
        </section>
        <section v-else class="card">
          <p class="empty">{{ t('resource.detail.noEvents') }}</p>
        </section>
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

.card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 1rem;
}

.cond-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0;
  border-bottom: 1px dashed var(--color-border);
}

.cond-row:last-child {
  border-bottom: 0;
}

.cond-type {
  width: 110px;
  font-size: 0.85rem;
  font-weight: 500;
}

.cond-reason {
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

.cond-age {
  margin-left: auto;
  font-size: 0.78rem;
  color: var(--color-text-muted);
  font-family: var(--font-mono);
}

.empty {
  margin: 0;
  text-align: center;
  color: var(--color-text-muted);
  font-size: 0.875rem;
  padding: 1rem;
}
</style>
