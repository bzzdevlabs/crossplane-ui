<script setup lang="ts">
import { computed } from 'vue';

import FieldRow from './FieldRow.vue';
import { getPath, setPath, useStringField, type Obj } from './path';

const props = defineProps<{ modelValue: Obj }>();
const emit = defineEmits<{ (e: 'update:modelValue', v: Obj): void }>();

const bind = (path: readonly string[]) =>
  useStringField(() => props.modelValue, (v) => emit('update:modelValue', v), path);

const name = bind(['metadata', 'name']);
const group = bind(['spec', 'group']);
const kind = bind(['spec', 'names', 'kind']);
const plural = bind(['spec', 'names', 'plural']);
const scope = bind(['spec', 'scope']);

// The XRD form edits the first version entry only. Multi-version XRDs fall
// back to the YAML view; the form keeps the invariant that versions[0] is the
// "active" version so simple XRDs remain editable end-to-end.
const versionName = computed({
  get: () => (getPath(props.modelValue, ['spec', 'versions', '0', 'name']) as string | undefined) ?? '',
  set: (value: string) => {
    const next = { ...props.modelValue };
    const versions = (getPath(next, ['spec', 'versions']) as Obj[] | undefined) ?? [];
    const first = { ...(versions[0] ?? {}), name: value };
    const copy = [...versions];
    copy[0] = first;
    setPath(next, ['spec', 'versions'], copy);
    emit('update:modelValue', next);
  },
});
</script>

<template>
  <div class="form-grid">
    <FieldRow label="metadata.name" hint="Must equal plural.group." required>
      <input v-model="name" type="text" placeholder="xexamples.example.org" />
    </FieldRow>
    <FieldRow label="spec.group" required>
      <input v-model="group" type="text" placeholder="example.org" />
    </FieldRow>
    <FieldRow label="spec.names.kind" required>
      <input v-model="kind" type="text" placeholder="XExample" />
    </FieldRow>
    <FieldRow label="spec.names.plural" required>
      <input v-model="plural" type="text" placeholder="xexamples" />
    </FieldRow>
    <FieldRow label="spec.scope" hint="v2 defaults XRs to Namespaced.">
      <select v-model="scope">
        <option value="">(default)</option>
        <option value="Namespaced">Namespaced</option>
        <option value="Cluster">Cluster</option>
        <option value="LegacyCluster">LegacyCluster</option>
      </select>
    </FieldRow>
    <FieldRow label="spec.versions[0].name" required>
      <input v-model="versionName" type="text" placeholder="v1alpha1" />
    </FieldRow>
  </div>
</template>

<style scoped>
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 0.75rem 1rem;
}
</style>
