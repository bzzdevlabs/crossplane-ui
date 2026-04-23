<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import {
  ApiError,
  applyConnector,
  getConnector,
  writeConnectorSecret,
  type ConnectorCR,
  type ConnectorSecretInjection,
  type ConnectorType,
} from '@/services/api';
import { loadGatewayConfig } from '@/services/config';
import {
  CONNECTOR_TEMPLATES,
  getAtPath,
  setAtPath,
  templateFor,
  type ConnectorField,
  type ConnectorTemplate,
} from '@/resources/connectorTemplates';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const isEdit = computed(() => route.name === 'connector-detail');
const existingName = computed(() =>
  typeof route.params.id === 'string' ? route.params.id : undefined,
);

const loading = ref(false);
const saving = ref(false);
const error = ref<string | null>(null);

interface FormState {
  id: string;
  type: ConnectorType;
  displayName: string;
  disabled: boolean;
  config: Record<string, unknown>;
  // plaintextSecrets keys are field `key`s from the template (nested paths ok).
  plaintextSecrets: Record<string, string>;
  // existingSecretRefs preserved when editing so we do not drop refs the user
  // did not touch (e.g. they did not re-enter the client secret).
  existingSecretRefs: ConnectorSecretInjection[];
}

const form = reactive<FormState>({
  id: '',
  type: 'oidc',
  displayName: '',
  disabled: false,
  config: {},
  plaintextSecrets: {},
  existingSecretRefs: [],
});

const authNamespace = ref('crossplane-ui');

const template = computed<ConnectorTemplate | undefined>(() => templateFor(form.type));

watch(
  () => form.type,
  (nt, ot) => {
    if (nt === ot) return;
    const tpl = templateFor(nt);
    if (!tpl) return;
    // On type change in create mode, reset config to the template defaults.
    if (!isEdit.value) {
      form.config = JSON.parse(JSON.stringify(tpl.defaults)) as Record<string, unknown>;
      form.plaintextSecrets = {};
    }
  },
);

function hydrateFromCR(cr: ConnectorCR) {
  form.id = cr.spec.id;
  form.type = cr.spec.type;
  form.displayName = cr.spec.name;
  form.disabled = cr.spec.disabled ?? false;
  form.config = JSON.parse(JSON.stringify(cr.spec.config ?? {})) as Record<string, unknown>;
  form.existingSecretRefs = [...(cr.spec.secretRefs ?? [])];
}

async function loadExisting() {
  if (!existingName.value) return;
  loading.value = true;
  try {
    const cr = await getConnector(existingName.value);
    hydrateFromCR(cr);
  } catch (e) {
    error.value = e instanceof ApiError ? `${e.code} (${e.status})` : String(e);
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  const cfg = await loadGatewayConfig();
  if (cfg.authNamespace) authNamespace.value = cfg.authNamespace;
  if (isEdit.value) {
    await loadExisting();
  } else {
    const tpl = template.value;
    if (tpl) {
      form.config = JSON.parse(JSON.stringify(tpl.defaults)) as Record<string, unknown>;
    }
  }
});

function fieldValue(f: ConnectorField): unknown {
  if (f.kind === 'secret') {
    return form.plaintextSecrets[f.key] ?? '';
  }
  return getAtPath(form.config, f.key);
}

function setFieldValue(f: ConnectorField, val: unknown) {
  if (f.kind === 'secret') {
    form.plaintextSecrets[f.key] = typeof val === 'string' ? val : '';
    return;
  }
  if (f.kind === 'bool') {
    setAtPath(form.config, f.key, Boolean(val));
    return;
  }
  if (f.kind === 'stringList') {
    const s = typeof val === 'string' ? val : '';
    const parts = s
      .split(',')
      .map((x) => x.trim())
      .filter(Boolean);
    setAtPath(form.config, f.key, parts);
    return;
  }
  setAtPath(form.config, f.key, typeof val === 'string' ? val : '');
}

function stringListAsText(f: ConnectorField): string {
  const v = getAtPath(form.config, f.key);
  if (Array.isArray(v)) return v.join(', ');
  return typeof v === 'string' ? v : '';
}

function secretBound(f: ConnectorField): boolean {
  if (f.kind !== 'secret') return false;
  return form.existingSecretRefs.some((r) => r.path === f.path);
}

async function save() {
  if (!template.value) return;
  saving.value = true;
  error.value = null;
  try {
    // Write any plaintext secrets first, and build the resulting secretRefs.
    const newRefs: ConnectorSecretInjection[] = [];
    const secretName = `connector-${form.id}`;
    const secretData: Record<string, string> = {};
    for (const f of template.value.fields) {
      if (f.kind !== 'secret') continue;
      const plaintext = form.plaintextSecrets[f.key];
      if (plaintext !== undefined && plaintext !== '') {
        secretData[f.secretKey] = plaintext;
        newRefs.push({ path: f.path, secretRef: { name: secretName, key: f.secretKey } });
      }
    }
    if (Object.keys(secretData).length > 0) {
      await writeConnectorSecret({
        namespace: authNamespace.value,
        name: secretName,
        data: secretData,
      });
    }

    // Merge: keep existing refs the user did not overwrite.
    const mergedRefs = [
      ...form.existingSecretRefs.filter((r) => !newRefs.some((n) => n.path === r.path)),
      ...newRefs,
    ];

    const body = {
      apiVersion: 'auth.crossplane-ui.io/v1alpha1',
      kind: 'Connector',
      metadata: { name: form.id },
      spec: {
        id: form.id,
        type: form.type,
        name: form.displayName,
        config: form.config,
        secretRefs: mergedRefs,
        disabled: form.disabled,
      },
    };
    await applyConnector(body);
    void router.push({ name: 'connectors' });
  } catch (e) {
    error.value = e instanceof ApiError ? `${e.code} (${e.status})` : String(e);
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <section class="form">
    <header class="head">
      <div>
        <h1>
          {{ isEdit ? t('connectors.edit') : t('connectors.create') }}
        </h1>
        <p v-if="template" class="muted">{{ template.description }}</p>
      </div>
      <router-link :to="{ name: 'connectors' }" class="btn">{{ t('common.back') }}</router-link>
    </header>

    <p v-if="loading" class="muted">{{ t('common.loading') }}</p>

    <div class="grid">
      <label>
        <span>{{ t('connectors.columns.id') }}</span>
        <input
          v-model="form.id"
          :disabled="isEdit"
          required
          pattern="^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
          placeholder="github-corp"
        />
        <small class="help">{{ t('connectors.help.id') }}</small>
      </label>

      <label>
        <span>{{ t('connectors.columns.type') }}</span>
        <select v-model="form.type" :disabled="isEdit">
          <option v-for="tpl in CONNECTOR_TEMPLATES" :key="tpl.type" :value="tpl.type">
            {{ tpl.label }}
          </option>
        </select>
      </label>

      <label>
        <span>{{ t('connectors.columns.name') }}</span>
        <input v-model="form.displayName" required placeholder="Corporate GitHub" />
      </label>

      <label class="row-checkbox">
        <input v-model="form.disabled" type="checkbox" />
        <span>{{ t('connectors.disabled') }}</span>
      </label>
    </div>

    <template v-if="template">
      <h2>{{ t('connectors.sections.config') }}</h2>
      <p v-if="template.docsHref" class="muted">
        <a :href="template.docsHref" target="_blank" rel="noopener">{{ t('connectors.docs') }} →</a>
      </p>

      <div class="grid">
        <label v-for="f in template.fields" :key="f.key">
          <span>
            {{ f.label }}
            <em v-if="f.required" class="req">*</em>
          </span>

          <input
            v-if="f.kind === 'string'"
            :value="fieldValue(f) as string"
            :placeholder="f.placeholder"
            :required="f.required"
            @input="(e) => setFieldValue(f, (e.target as HTMLInputElement).value)"
          />

          <label v-else-if="f.kind === 'bool'" class="row-checkbox inner">
            <input
              type="checkbox"
              :checked="Boolean(fieldValue(f))"
              @change="(e) => setFieldValue(f, (e.target as HTMLInputElement).checked)"
            />
          </label>

          <input
            v-else-if="f.kind === 'stringList'"
            :value="stringListAsText(f)"
            :placeholder="f.placeholder"
            @input="(e) => setFieldValue(f, (e.target as HTMLInputElement).value)"
          />

          <input
            v-else-if="f.kind === 'secret'"
            type="password"
            :value="fieldValue(f) as string"
            :placeholder="secretBound(f) ? t('connectors.help.secretSet') : f.placeholder"
            @input="(e) => setFieldValue(f, (e.target as HTMLInputElement).value)"
          />

          <small v-if="f.helpText" class="help">{{ f.helpText }}</small>
        </label>
      </div>
    </template>

    <p v-if="error" class="error">{{ error }}</p>

    <div class="actions">
      <button type="button" class="btn primary" :disabled="saving" @click="save">
        {{ saving ? t('connectors.saving') : t('connectors.save') }}
      </button>
    </div>
  </section>
</template>

<style scoped>
.form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: 48rem;
}

.head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

h1 {
  margin: 0;
  font-size: 1.5rem;
}

h2 {
  margin: 1rem 0 0.5rem;
  font-size: 1.1rem;
}

.muted {
  color: var(--color-text-muted);
  margin: 0;
}

.error {
  color: var(--color-danger, #c0392b);
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr));
  gap: 0.75rem 1rem;
}

label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.9rem;
}

label span {
  font-weight: 500;
}

input,
select {
  padding: 0.4rem 0.5rem;
  border: 1px solid var(--color-border, #ccd);
  border-radius: 4px;
  font: inherit;
}

.row-checkbox {
  flex-direction: row;
  align-items: center;
  gap: 0.5rem;
}

.row-checkbox.inner {
  align-self: start;
}

.help {
  color: var(--color-text-muted);
  font-size: 0.8rem;
}

.req {
  color: var(--color-danger, #c0392b);
  font-style: normal;
}

.actions {
  display: flex;
  justify-content: flex-end;
}

.btn {
  display: inline-block;
  padding: 0.45rem 0.9rem;
  border-radius: 4px;
  border: 1px solid transparent;
  cursor: pointer;
  background: transparent;
  color: inherit;
  font: inherit;
  text-decoration: none;
}

.btn.primary {
  background: var(--color-accent, #2e4fd9);
  color: #fff;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
