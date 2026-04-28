<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter, type RouteLocationRaw } from 'vue-router';

import FormShell from '@/components/forms/FormShell.vue';
import { FORM_SCHEMAS, type FormSchema } from '@/components/forms/schemas';
import type { Obj } from '@/components/forms/path';
import BreadcrumbBar from '@/components/ui/BreadcrumbBar.vue';
import { applyResource, type ResourceRef } from '@/services/api';

const { t } = useI18n();
const router = useRouter();
const route = useRoute();

function initialSchema(): FormSchema {
  const requested = route.query.kind;
  const id = Array.isArray(requested) ? requested[0] : requested;
  if (id) {
    const match = FORM_SCHEMAS.find((s) => s.id === id);
    if (match) return match;
  }
  return FORM_SCHEMAS[0]!;
}

const selected = ref<FormSchema>(initialSchema());
const object = ref<Obj>(selected.value.skeleton());
const saving = ref(false);
const errorMsg = ref<string | null>(null);

watch(selected, (s) => {
  object.value = s.skeleton();
  errorMsg.value = null;
});

function pluralFromKind(kind: string): string {
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

const metaText = computed(() => {
  const apiVersion = typeof object.value.apiVersion === 'string' ? object.value.apiVersion : '';
  const kind = typeof object.value.kind === 'string' ? object.value.kind : '';
  return [apiVersion, kind].filter(Boolean).join(' · ');
});

const breadcrumbs = computed(() => [
  { label: t('products.crossplane.label'), to: { name: 'crossplane-dashboard' } as RouteLocationRaw },
  { label: t('resource.create') },
]);

async function apply(): Promise<void> {
  if (!canApply.value) return;
  saving.value = true;
  errorMsg.value = null;
  try {
    const ref = targetRef();
    await applyResource(ref, object.value);
    await router.replace({
      name: 'resource-detail',
      params: { resource: ref.resource, name: ref.name },
      query: {
        ...(ref.group ? { group: ref.group } : {}),
        ...(ref.version ? { version: ref.version } : {}),
        ...(ref.namespace ? { namespace: ref.namespace } : {}),
      },
    });
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err);
  } finally {
    saving.value = false;
  }
}

function goBack(): void {
  void router.push({ name: 'crossplane-dashboard' });
}
</script>

<template>
  <section class="create">
    <BreadcrumbBar :items="breadcrumbs" />

    <header class="page-header">
      <div>
        <h1>{{ t('resource.create') }}</h1>
        <p class="muted">{{ t('resource.createHint') }}</p>
      </div>
      <label class="picker">
        {{ t('resource.template') }}
        <select v-model="selected">
          <option v-for="tmpl in FORM_SCHEMAS" :key="tmpl.id" :value="tmpl">
            {{ tmpl.label }}
          </option>
        </select>
      </label>
    </header>

    <p v-if="errorMsg" class="error">{{ errorMsg }}</p>

    <section class="card">
      <div class="card-body">
        <FormShell v-model="object" :form-component="selected.component">
          <template #meta>
            <span class="meta">{{ metaText }}</span>
          </template>
        </FormShell>
      </div>
      <footer class="card-footer">
        <button type="button" class="btn neutral" @click="goBack">
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          class="btn primary"
          :disabled="saving || !canApply"
          @click="apply"
        >
          {{ saving ? t('resource.saving') : t('resource.apply') }}
        </button>
      </footer>
    </section>
  </section>
</template>

<style scoped>
.create {
  display: flex;
  flex-direction: column;
  gap: 1rem;
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
  font-size: 1.35rem;
  font-weight: 600;
}

.muted {
  margin: 0.2rem 0 0;
  color: var(--color-text-muted);
  font-size: 0.9rem;
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

.error {
  margin: 0;
  color: var(--color-danger);
  white-space: pre-wrap;
}

.card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.card-body {
  padding: 0.75rem;
}

.card-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0.75rem;
  border-top: 1px solid var(--color-border);
}

.meta {
  margin-left: auto;
  font-size: 0.78rem;
  color: var(--color-text-muted);
  font-family: var(--font-mono);
}

.btn {
  padding: 0.4rem 0.9rem;
  border-radius: 6px;
  font: inherit;
  font-size: 0.875rem;
  cursor: pointer;
  border: 1px solid var(--color-border);
}

.btn.neutral {
  background: var(--color-surface);
  color: inherit;
}

.btn.primary {
  border-color: var(--color-accent);
  background: var(--color-accent);
  color: var(--color-on-accent);
}

.btn[disabled] {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
