<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router';

export interface Crumb {
  readonly label: string;
  readonly to?: RouteLocationRaw;
}

defineProps<{
  items: readonly Crumb[];
}>();
</script>

<template>
  <nav class="crumbs">
    <template v-for="(c, i) in items" :key="i">
      <RouterLink v-if="c.to" :to="c.to">{{ c.label }}</RouterLink>
      <span v-else class="current">{{ c.label }}</span>
      <span v-if="i < items.length - 1" class="sep">/</span>
    </template>
  </nav>
</template>

<style scoped>
.crumbs {
  display: flex;
  gap: 0.35rem;
  color: var(--color-text-muted);
  font-size: 0.85rem;
  flex-wrap: wrap;
}

a {
  color: inherit;
  text-decoration: none;
}

a:hover {
  color: var(--color-text);
}

.current {
  color: var(--color-text);
}

.sep {
  opacity: 0.5;
}
</style>
