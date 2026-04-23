<script setup lang="ts">
// Sticky action bar pinned to the viewport bottom. Used by resource forms
// (Cancel / Edit as YAML / Apply) — see ADR-0007.
</script>

<template>
  <div class="footer">
    <div class="left">
      <slot name="left" />
    </div>
    <div class="right">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.footer {
  position: sticky;
  bottom: 0;
  z-index: 5;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.75rem 1.25rem;
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
  margin: 1.5rem -1.75rem -1.5rem;
}

.left,
.right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
</style>
