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
const runtimeConfigName = bind(['spec', 'runtimeConfigRef', 'name']);
</script>

<template>
  <div class="form-grid">
    <FieldRow label="metadata.name" required>
      <input v-model="name" type="text" placeholder="provider-kubernetes" />
    </FieldRow>
    <FieldRow label="spec.package" required hint="OCI reference to the provider package.">
      <input
        v-model="pkg"
        type="text"
        placeholder="xpkg.upbound.io/crossplane-contrib/provider-kubernetes:v0.16.0"
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
    <FieldRow label="spec.runtimeConfigRef.name" hint="Leave blank to use the default runtime.">
      <input v-model="runtimeConfigName" type="text" placeholder="default" />
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
