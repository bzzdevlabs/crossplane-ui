<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import FormShell from '@/components/forms/FormShell.vue';
import { FORM_SCHEMAS, type FormSchema } from '@/components/forms/schemas';
import type { Obj } from '@/components/forms/path';
import { applyResource, type ResourceRef } from '@/services/api';

const { t } = useI18n();
const router = useRouter();

const selected = ref<FormSchema>(FORM_SCHEMAS[0]!);
const object = ref<Obj>(FORM_SCHEMAS[0]!.skeleton());
const saving = ref(false);
const error = ref<string | null>(null);

watch(selected, (s) => {
  object.value = s.skeleton();
  error.value = null;
});

function pluralFromKind(kind: string): string {
  // Best-effort English plural for ProviderConfig-style kinds the user types
  // themselves. Crossplane CRDs in the wild use lowercased plurals; we
  // mimic the kubebuilder default.
  const lower = kind.toLowerCase();
  if (lower.endsWith('s')) return lower;
  if (lower.endsWith('y')) return `${lower.slice(0, -1)}ies`;
  return `${lower}s`;
}

function targetRef(): ResourceRef {
  const obj = object.value;
  const meta = (obj.metadata ?? {}) as Obj;
  const name = typeof meta.name === 'string' ? meta.name : '';
  if (!name) throw new Error('metadata.name is required');

  if (selected.value.ref.resource !== '') {
    return {
      ...selected.value.ref,
      namespace: typeof meta.namespace === 'string' ? meta.namespace : undefined,
      name,
    };
  }

  // Dynamic GVR (ProviderConfig and other provider-supplied kinds): derive
  // group/version from the object's apiVersion and plural from kind.
  const apiVersion = typeof obj.apiVersion === 'string' ? obj.apiVersion : '';
  const kind = typeof obj.kind === 'string' ? obj.kind : '';
  if (!apiVersion || !kind) {
    throw new Error('apiVersion and kind are required');
  }
  const [group, version] = apiVersion.includes('/')
    ? apiVersion.split('/')
    : ['', apiVersion];
  return {
    group: group ?? '',
    version: version ?? '',
    resource: pluralFromKind(kind),
    namespace: typeof meta.namespace === 'string' ? meta.namespace : undefined,
    name,
  };
}

const canApply = computed(() => !saving.value);

async function apply(): Promise<void> {
  if (!canApply.value) return;
  saving.value = true;
  error.value = null;
  try {
    const ref = targetRef();
    await applyResource(ref, object.value);
    await router.replace({
      name: 'resource-detail',
      params: { group: ref.group, version: ref.version, resource: ref.resource, name: ref.name },
      query: ref.namespace ? { namespace: ref.namespace } : undefined,
    });
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <section class="create">
    <nav class="breadcrumbs">
      <RouterLink :to="{ name: 'home' }">{{ t('nav.home') }}</RouterLink>
      <span>/</span>
      <span class="current">{{ t('resource.create') }}</span>
    </nav>

    <header class="page-header">
      <div>
        <h1>{{ t('resource.create') }}</h1>
        <p class="muted">{{ t('resource.createHint') }}</p>
      </div>
      <div class="actions">
        <label class="picker">
          {{ t('resource.template') }}
          <select v-model="selected">
            <option v-for="tmpl in FORM_SCHEMAS" :key="tmpl.id" :value="tmpl">
              {{ tmpl.label }}
            </option>
          </select>
        </label>
        <button type="button" class="primary" :disabled="!canApply" @click="apply">
          {{ saving ? t('resource.saving') : t('resource.apply') }}
        </button>
      </div>
    </header>

    <p v-if="error" class="error">{{ error }}</p>

    <FormShell v-model="object" :form-component="selected.component" />
  </section>
</template>

<style scoped>
.create {
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
  flex-wrap: wrap;
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
  align-items: center;
  gap: 0.75rem;
}

.picker {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: var(--color-text-muted);
}

.picker select {
  padding: 0.3rem 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: inherit;
  font: inherit;
}

button.primary {
  padding: 0.45rem 1rem;
  border: 1px solid var(--color-accent);
  background: var(--color-accent);
  color: var(--color-on-accent);
  border-radius: 6px;
  font: inherit;
  cursor: pointer;
}

button.primary[disabled] {
  opacity: 0.5;
  cursor: not-allowed;
}

.error {
  margin: 0;
  color: var(--color-danger);
  white-space: pre-wrap;
}
</style>
