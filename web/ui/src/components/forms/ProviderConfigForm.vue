<script setup lang="ts">
import { computed } from 'vue';

import FieldRow from './FieldRow.vue';
import { getPath, useStringField, type Obj } from './path';

const props = defineProps<{ modelValue: Obj }>();
const emit = defineEmits<{ (e: 'update:modelValue', v: Obj): void }>();

const bind = (path: readonly string[]) =>
  useStringField(() => props.modelValue, (v) => emit('update:modelValue', v), path);

const apiVersion = bind(['apiVersion']);
const kind = bind(['kind']);
const name = bind(['metadata', 'name']);
const namespace = bind(['metadata', 'namespace']);
const credentialsSource = bind(['spec', 'credentials', 'source']);
const secretNamespace = bind(['spec', 'credentials', 'secretRef', 'namespace']);
const secretName = bind(['spec', 'credentials', 'secretRef', 'name']);
const secretKey = bind(['spec', 'credentials', 'secretRef', 'key']);

const showSecretFields = computed(
  () => (getPath(props.modelValue, ['spec', 'credentials', 'source']) as string) === 'Secret',
);
</script>

<template>
  <div class="form-grid">
    <FieldRow
      label="apiVersion"
      required
      hint="ProviderConfigs are defined by each provider; set the correct apiVersion."
    >
      <input v-model="apiVersion" type="text" placeholder="aws.upbound.io/v1beta1" />
    </FieldRow>
    <FieldRow label="kind" required>
      <input v-model="kind" type="text" placeholder="ProviderConfig" />
    </FieldRow>
    <FieldRow label="metadata.name" required>
      <input v-model="name" type="text" placeholder="default" />
    </FieldRow>
    <FieldRow label="metadata.namespace" hint="Leave blank for cluster-scoped ProviderConfigs.">
      <input v-model="namespace" type="text" />
    </FieldRow>
    <FieldRow label="spec.credentials.source">
      <select v-model="credentialsSource">
        <option value="">(none)</option>
        <option value="Secret">Secret</option>
        <option value="InjectedIdentity">InjectedIdentity</option>
        <option value="Environment">Environment</option>
        <option value="Filesystem">Filesystem</option>
        <option value="Upbound">Upbound</option>
      </select>
    </FieldRow>
    <template v-if="showSecretFields">
      <FieldRow label="spec.credentials.secretRef.namespace" required>
        <input v-model="secretNamespace" type="text" placeholder="crossplane-system" />
      </FieldRow>
      <FieldRow label="spec.credentials.secretRef.name" required>
        <input v-model="secretName" type="text" placeholder="cloud-credentials" />
      </FieldRow>
      <FieldRow label="spec.credentials.secretRef.key" required>
        <input v-model="secretKey" type="text" placeholder="creds" />
      </FieldRow>
    </template>
  </div>
</template>

<style scoped>
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 0.75rem 1rem;
}
</style>
