<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { parse as parseYaml, stringify as stringifyYaml } from 'yaml';

import YamlEditor from '@/components/YamlEditor.vue';
import {
  applyResource,
  deleteResource,
  getResource,
  type ResourceRef,
} from '@/services/api';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const loading = ref(false);
const saving = ref(false);
const error = ref<string | null>(null);
const notice = ref<string | null>(null);
const original = ref<string>('');
const draft = ref<string>('');

const ref_ = computed<ResourceRef>(() => {
  const params = route.params;
  return {
    group: String(params.group ?? ''),
    version: String(params.version ?? ''),
    resource: String(params.resource ?? ''),
    name: String(params.name ?? ''),
    namespace: (route.query.namespace as string) || undefined,
  };
});

const dirty = computed(() => draft.value !== original.value);

function stripServerFields(obj: Record<string, unknown>): Record<string, unknown> {
  const clone = structuredClone(obj);
  const meta = (clone.metadata ?? {}) as Record<string, unknown>;
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
  error.value = null;
  notice.value = null;
  try {
    const obj = await getResource<Record<string, unknown>>(ref_.value);
    const editable = stripServerFields(obj);
    const yaml = stringifyYaml(editable, { indent: 2, lineWidth: 0 });
    original.value = yaml;
    draft.value = yaml;
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

async function save(): Promise<void> {
  if (saving.value) return;
  saving.value = true;
  error.value = null;
  notice.value = null;
  try {
    const parsed = parseYaml(draft.value) as Record<string, unknown>;
    const applied = (await applyResource<Record<string, unknown>>(ref_.value, parsed));
    const editable = stripServerFields(applied);
    const yaml = stringifyYaml(editable, { indent: 2, lineWidth: 0 });
    original.value = yaml;
    draft.value = yaml;
    notice.value = t('resource.saved');
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    saving.value = false;
  }
}

async function remove(): Promise<void> {
  if (!window.confirm(t('resource.confirmDelete', { name: ref_.value.name }))) {
    return;
  }
  saving.value = true;
  error.value = null;
  try {
    await deleteResource(ref_.value);
    await router.replace({ name: 'home' });
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
    saving.value = false;
  }
}

watch(ref_, load, { immediate: false });
onMounted(load);
</script>

<template>
  <section class="detail">
    <nav class="breadcrumbs">
      <RouterLink :to="{ name: 'home' }">{{ t('nav.home') }}</RouterLink>
      <span>/</span>
      <span>{{ ref_.resource }}</span>
      <span>/</span>
      <span class="current">{{ ref_.name }}</span>
    </nav>

    <header class="page-header">
      <div>
        <h1>{{ ref_.name }}</h1>
        <p class="muted">{{ ref_.group || 'core' }}/{{ ref_.version }} · {{ ref_.resource }}</p>
      </div>
      <div class="actions">
        <button type="button" :disabled="saving" @click="load">{{ t('home.refresh') }}</button>
        <button type="button" class="danger" :disabled="saving" @click="remove">
          {{ t('common.delete') }}
        </button>
        <button
          type="button"
          class="primary"
          :disabled="saving || !dirty"
          @click="save"
        >
          {{ saving ? t('resource.saving') : t('resource.apply') }}
        </button>
      </div>
    </header>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="notice" class="notice">{{ notice }}</p>

    <p v-if="loading" class="muted">{{ t('home.loading') }}</p>
    <YamlEditor v-else v-model="draft" />
  </section>
</template>

<style scoped>
.detail {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.breadcrumbs {
  display: flex;
  gap: 0.35rem;
  color: var(--color-text-muted);
  font-size: 0.85rem;
}

.breadcrumbs a {
  color: inherit;
  text-decoration: none;
}

.breadcrumbs .current {
  color: var(--color-text);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 1rem;
}

h1 {
  margin: 0;
  font-size: 1.3rem;
}

.muted {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 0.9rem;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.actions button {
  padding: 0.4rem 0.9rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.actions button[disabled] {
  opacity: 0.5;
  cursor: not-allowed;
}

.actions .primary {
  border-color: var(--color-accent);
  background: var(--color-accent);
  color: var(--color-on-accent);
}

.actions .danger {
  border-color: var(--color-danger);
  color: var(--color-danger);
  background: transparent;
}

.error {
  color: var(--color-danger);
  margin: 0;
  white-space: pre-wrap;
}

.notice {
  color: #1f7a3a;
  margin: 0;
}
</style>
