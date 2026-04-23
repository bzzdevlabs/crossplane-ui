<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import StickyFooter from '@/components/ui/StickyFooter.vue';

defineProps<{
  saving?: boolean;
  canApply?: boolean;
  applyLabel?: string;
  showCancel?: boolean;
}>();

defineEmits<{
  (e: 'cancel'): void;
  (e: 'apply'): void;
}>();

const { t } = useI18n();
</script>

<template>
  <div class="form-template">
    <div class="body">
      <slot />
    </div>
    <StickyFooter>
      <template #left>
        <slot name="footerLeft" />
      </template>
      <button
        v-if="showCancel !== false"
        type="button"
        class="secondary"
        @click="$emit('cancel')"
      >
        {{ t('common.cancel') }}
      </button>
      <slot name="footerActions" />
      <button
        type="button"
        class="primary"
        :disabled="saving || canApply === false"
        @click="$emit('apply')"
      >
        {{ saving ? t('resource.saving') : applyLabel ?? t('resource.apply') }}
      </button>
    </StickyFooter>
  </div>
</template>

<style scoped>
.form-template {
  display: flex;
  flex-direction: column;
  min-height: 100%;
}

.body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

button {
  padding: 0.4rem 1rem;
  border-radius: 6px;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: inherit;
  font: inherit;
  cursor: pointer;
}

button[disabled] {
  opacity: 0.5;
  cursor: not-allowed;
}

button.primary {
  border-color: var(--color-accent);
  background: var(--color-accent);
  color: var(--color-on-accent);
}

button.secondary {
  border-color: var(--color-border);
}
</style>
