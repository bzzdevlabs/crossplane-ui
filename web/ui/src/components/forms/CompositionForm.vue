<script setup lang="ts">
import FieldRow from './FieldRow.vue';
import { useStringField, type Obj } from './path';

const props = defineProps<{ modelValue: Obj }>();
const emit = defineEmits<{ (e: 'update:modelValue', v: Obj): void }>();

const bind = (path: readonly string[]) =>
  useStringField(() => props.modelValue, (v) => emit('update:modelValue', v), path);

const name = bind(['metadata', 'name']);
const xrApiVersion = bind(['spec', 'compositeTypeRef', 'apiVersion']);
const xrKind = bind(['spec', 'compositeTypeRef', 'kind']);
const mode = bind(['spec', 'mode']);
</script>

<template>
  <div class="form-grid">
    <FieldRow label="metadata.name" required>
      <input v-model="name" type="text" placeholder="example-composition" />
    </FieldRow>
    <FieldRow label="spec.compositeTypeRef.apiVersion" required>
      <input v-model="xrApiVersion" type="text" placeholder="example.org/v1alpha1" />
    </FieldRow>
    <FieldRow label="spec.compositeTypeRef.kind" required>
      <input v-model="xrKind" type="text" placeholder="XExample" />
    </FieldRow>
    <FieldRow label="spec.mode" hint="Pipeline is the v2 default; Resources is legacy.">
      <select v-model="mode">
        <option value="Pipeline">Pipeline</option>
        <option value="Resources">Resources</option>
      </select>
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
