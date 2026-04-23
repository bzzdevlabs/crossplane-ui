<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { parse as parseYaml, stringify as stringifyYaml } from 'yaml';

import YamlEditor from '@/components/YamlEditor.vue';
import { applyResource, type ResourceRef } from '@/services/api';

interface Template {
  readonly id: string;
  readonly label: string;
  readonly ref: Omit<ResourceRef, 'name'>;
  readonly skeleton: Record<string, unknown>;
}

const TEMPLATES: readonly Template[] = [
  {
    id: 'composition',
    label: 'Composition',
    ref: { group: 'apiextensions.crossplane.io', version: 'v1', resource: 'compositions' },
    skeleton: {
      apiVersion: 'apiextensions.crossplane.io/v1',
      kind: 'Composition',
      metadata: { name: 'example-composition' },
      spec: {
        compositeTypeRef: { apiVersion: 'example.org/v1alpha1', kind: 'XExample' },
        mode: 'Pipeline',
        pipeline: [],
      },
    },
  },
  {
    id: 'provider',
    label: 'Provider',
    ref: { group: 'pkg.crossplane.io', version: 'v1', resource: 'providers' },
    skeleton: {
      apiVersion: 'pkg.crossplane.io/v1',
      kind: 'Provider',
      metadata: { name: 'example-provider' },
      spec: { package: 'xpkg.upbound.io/example/provider:v1.0.0' },
    },
  },
  {
    id: 'function',
    label: 'Function',
    ref: { group: 'pkg.crossplane.io', version: 'v1', resource: 'functions' },
    skeleton: {
      apiVersion: 'pkg.crossplane.io/v1',
      kind: 'Function',
      metadata: { name: 'example-function' },
      spec: { package: 'xpkg.upbound.io/crossplane-contrib/function-go-templating:v0.9.0' },
    },
  },
];

const { t } = useI18n();
const router = useRouter();

const selected = ref<Template>(TEMPLATES[0]!);
const draft = ref<string>(stringifyYaml(TEMPLATES[0]!.skeleton, { indent: 2 }));
const saving = ref(false);
const error = ref<string | null>(null);

watch(selected, (tmpl) => {
  draft.value = stringifyYaml(tmpl.skeleton, { indent: 2 });
  error.value = null;
});

const canApply = computed(() => !saving.value && draft.value.trim().length > 0);

async function apply(): Promise<void> {
  if (!canApply.value) return;
  saving.value = true;
  error.value = null;
  try {
    const parsed = parseYaml(draft.value) as Record<string, unknown>;
    const meta = (parsed.metadata ?? {}) as Record<string, unknown>;
    const name = typeof meta.name === 'string' ? meta.name : '';
    if (!name) {
      throw new Error('metadata.name is required');
    }
    const ref: ResourceRef = { ...selected.value.ref, name };
    await applyResource(ref, parsed);
    await router.replace({
      name: 'resource-detail',
      params: { ...selected.value.ref, name },
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
            <option v-for="tmpl in TEMPLATES" :key="tmpl.id" :value="tmpl">
              {{ tmpl.label }}
            </option>
          </select>
        </label>
        <button
          type="button"
          class="primary"
          :disabled="!canApply"
          @click="apply"
        >
          {{ saving ? t('resource.saving') : t('resource.apply') }}
        </button>
      </div>
    </header>

    <p v-if="error" class="error">{{ error }}</p>

    <YamlEditor v-model="draft" />
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
