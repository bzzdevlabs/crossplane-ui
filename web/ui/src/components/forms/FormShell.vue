<script setup lang="ts">
import { ref, watch, type Component } from 'vue';
import { useI18n } from 'vue-i18n';
import { parse as parseYaml, stringify as stringifyYaml } from 'yaml';

import YamlEditor from '@/components/YamlEditor.vue';

import type { Obj } from './path';

// FormShell wraps a kind-specific form with a YAML editor, toggled by a
// mode switch. The canonical state lives in `modelValue` (an Obj); the YAML
// buffer is derived on entry into yaml mode and parsed back out on exit or
// on edit. A parse error keeps the mode on `yaml` and surfaces the error
// instead of applying a corrupt object upstream.

const props = defineProps<{
  modelValue: Obj;
  formComponent?: Component | null;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: Obj): void;
}>();

const { t } = useI18n();

const mode = ref<'form' | 'yaml'>(props.formComponent ? 'form' : 'yaml');
const yamlDraft = ref<string>(stringifyYaml(props.modelValue, { indent: 2, lineWidth: 0 }));
const yamlError = ref<string | null>(null);

// Keep YAML buffer synced with upstream changes while the user is in form
// mode. In yaml mode the user is authoring text, so we leave the buffer
// alone and let onYamlChange flow updates the other way.
watch(
  () => props.modelValue,
  (next) => {
    if (mode.value === 'form') {
      yamlDraft.value = stringifyYaml(next, { indent: 2, lineWidth: 0 });
    }
  },
  { deep: true },
);

watch(
  () => props.formComponent,
  (cmp) => {
    if (!cmp && mode.value === 'form') mode.value = 'yaml';
  },
);

function switchMode(target: 'form' | 'yaml'): void {
  if (target === mode.value) return;
  if (target === 'form') {
    // User leaves yaml → commit the buffer by parsing once. If it fails,
    // block the switch and surface the error.
    try {
      const parsed = parseYaml(yamlDraft.value) as unknown;
      if (parsed == null || typeof parsed !== 'object' || Array.isArray(parsed)) {
        yamlError.value = 'YAML document must be an object.';
        return;
      }
      yamlError.value = null;
      emit('update:modelValue', parsed as Obj);
      mode.value = 'form';
    } catch (err) {
      yamlError.value = err instanceof Error ? err.message : String(err);
    }
    return;
  }
  // form → yaml: snapshot the current object as text for editing.
  yamlDraft.value = stringifyYaml(props.modelValue, { indent: 2, lineWidth: 0 });
  yamlError.value = null;
  mode.value = 'yaml';
}

function onYamlChange(next: string): void {
  yamlDraft.value = next;
  try {
    const parsed = parseYaml(next) as unknown;
    if (parsed == null || typeof parsed !== 'object' || Array.isArray(parsed)) {
      yamlError.value = 'YAML document must be an object.';
      return;
    }
    yamlError.value = null;
    emit('update:modelValue', parsed as Obj);
  } catch (err) {
    yamlError.value = err instanceof Error ? err.message : String(err);
  }
}

function onFormChange(next: Obj): void {
  emit('update:modelValue', next);
}
</script>

<template>
  <div class="shell">
    <div class="toolbar">
      <div class="modes" role="tablist">
        <button
          type="button"
          role="tab"
          :aria-selected="mode === 'form'"
          :disabled="!formComponent"
          :class="{ active: mode === 'form' }"
          @click="switchMode('form')"
        >
          {{ t('resource.modeForm') }}
        </button>
        <button
          type="button"
          role="tab"
          :aria-selected="mode === 'yaml'"
          :class="{ active: mode === 'yaml' }"
          @click="switchMode('yaml')"
        >
          {{ t('resource.modeYaml') }}
        </button>
      </div>
      <p v-if="yamlError" class="error">{{ yamlError }}</p>
    </div>

    <component
      :is="formComponent"
      v-if="mode === 'form' && formComponent"
      :model-value="modelValue"
      @update:model-value="onFormChange"
    />
    <YamlEditor
      v-else
      :model-value="yamlDraft"
      @update:model-value="onYamlChange"
    />
  </div>
</template>

<style scoped>
.shell {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.modes {
  display: inline-flex;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  overflow: hidden;
}

.modes button {
  padding: 0.35rem 0.9rem;
  border: 0;
  border-left: 1px solid var(--color-border);
  background: var(--color-surface);
  color: inherit;
  font: inherit;
  font-size: 0.85rem;
  cursor: pointer;
}

.modes button:first-child {
  border-left: 0;
}

.modes button.active {
  background: var(--color-accent);
  color: var(--color-on-accent);
}

.modes button[disabled] {
  opacity: 0.5;
  cursor: not-allowed;
}

.error {
  margin: 0;
  color: var(--color-danger);
  font-size: 0.8rem;
  white-space: pre-wrap;
}
</style>
