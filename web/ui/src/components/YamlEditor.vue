<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import * as monaco from 'monaco-editor';
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
import JsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';

// Wire monaco web workers through Vite's `?worker` suffix. Without this the
// editor logs "Could not create web worker(s)" and falls back to running all
// language services on the main thread.
self.MonacoEnvironment = {
  getWorker(_workerId: string, label: string) {
    if (label === 'json') return new JsonWorker();
    return new EditorWorker();
  },
};

const props = defineProps<{
  modelValue: string;
  readOnly?: boolean;
  height?: string;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void;
}>();

const container = ref<HTMLDivElement | null>(null);
let editor: monaco.editor.IStandaloneCodeEditor | null = null;

onMounted(() => {
  if (!container.value) return;
  editor = monaco.editor.create(container.value, {
    value: props.modelValue,
    language: 'yaml',
    theme: matchMedia('(prefers-color-scheme: dark)').matches ? 'vs-dark' : 'vs',
    readOnly: props.readOnly ?? false,
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
    automaticLayout: true,
    tabSize: 2,
    fontSize: 13,
    fontFamily:
      'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace',
  });
  editor.onDidChangeModelContent(() => {
    if (editor) emit('update:modelValue', editor.getValue());
  });
});

watch(
  () => props.modelValue,
  (next) => {
    if (editor && editor.getValue() !== next) {
      editor.setValue(next);
    }
  },
);

watch(
  () => props.readOnly,
  (ro) => {
    editor?.updateOptions({ readOnly: ro ?? false });
  },
);

onBeforeUnmount(() => {
  editor?.dispose();
  editor = null;
});
</script>

<template>
  <div ref="container" class="yaml-editor" :style="{ height: height ?? '60vh' }" />
</template>

<style scoped>
.yaml-editor {
  width: 100%;
  min-height: 240px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  overflow: hidden;
}
</style>
