<script setup lang="ts">
import type { Component } from 'vue';

defineProps<{
  icon?: Component;
  label?: string;
  color: string;
  size?: 'sm' | 'md' | 'lg';
}>();

const ICON_PX: Record<'sm' | 'md' | 'lg', number> = { sm: 14, md: 18, lg: 32 };
</script>

<template>
  <span
    class="badge"
    :data-size="size ?? 'md'"
    :style="{ background: color }"
    :title="label"
  >
    <component :is="icon" v-if="icon" :size="ICON_PX[size ?? 'md']" aria-hidden="true" />
    <template v-else>{{ label }}</template>
  </span>
</template>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 700;
  border-radius: 6px;
  letter-spacing: 0.02em;
  flex-shrink: 0;
}

.badge[data-size='md'] {
  width: 2rem;
  height: 2rem;
  font-size: 0.85rem;
}

.badge[data-size='sm'] {
  width: 1.5rem;
  height: 1.5rem;
  font-size: 0.7rem;
}

.badge[data-size='lg'] {
  width: 3.5rem;
  height: 3.5rem;
  font-size: 1.5rem;
  border-radius: 8px;
}
</style>
