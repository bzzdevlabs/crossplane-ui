<script setup lang="ts">
import FieldRow from './FieldRow.vue';
import { useStringField, type Obj } from './path';

const props = defineProps<{ modelValue: Obj }>();
const emit = defineEmits<{ (e: 'update:modelValue', v: Obj): void }>();

const bind = (path: readonly string[]) =>
  useStringField(() => props.modelValue, (v) => emit('update:modelValue', v), path);

const name = bind(['metadata', 'name']);
const pkg = bind(['spec', 'package']);
const pullPolicy = bind(['spec', 'packagePullPolicy']);
</script>

<template>
  <div class="form-grid">
    <FieldRow label="metadata.name" required>
      <input v-model="name" type="text" placeholder="function-go-templating" />
    </FieldRow>
    <FieldRow label="spec.package" required hint="OCI reference to the function package.">
      <input
        v-model="pkg"
        type="text"
        placeholder="xpkg.upbound.io/crossplane-contrib/function-go-templating:v0.9.0"
      />
    </FieldRow>
    <FieldRow label="spec.packagePullPolicy">
      <select v-model="pullPolicy">
        <option value="">(cluster default)</option>
        <option value="IfNotPresent">IfNotPresent</option>
        <option value="Always">Always</option>
        <option value="Never">Never</option>
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
