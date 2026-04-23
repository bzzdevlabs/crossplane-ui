<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import ProductBadge from '@/components/ui/ProductBadge.vue';
import type { Product } from '@/products';

defineProps<{ product: Product; active?: boolean }>();

const { t } = useI18n();
</script>

<template>
  <RouterLink class="link" :class="{ active }" :to="product.defaultRoute">
    <ProductBadge :icon="product.icon" :color="product.badgeColor" size="md" />
    <div class="info">
      <div class="label">{{ t(product.labelKey) }}</div>
      <div v-if="product.subtitleKey" class="sub">{{ t(product.subtitleKey) }}</div>
    </div>
  </RouterLink>
</template>

<style scoped>
.link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0.85rem;
  border-radius: 8px;
  color: inherit;
  text-decoration: none;
  border: 1px solid transparent;
}

.link:hover {
  background: var(--color-accent-subtle);
}

.link.active {
  background: var(--color-accent-subtle);
  border-color: var(--color-accent);
}

.info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.label {
  font-weight: 600;
  font-size: 0.95rem;
}

.sub {
  color: var(--color-text-muted);
  font-size: 0.8rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
