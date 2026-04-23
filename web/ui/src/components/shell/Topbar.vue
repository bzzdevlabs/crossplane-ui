<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { Menu } from 'lucide-vue-next';

import ProductBadge from '@/components/ui/ProductBadge.vue';
import { useUiStore } from '@/stores/ui';
import type { Product } from '@/products';

import NamespacePicker from './NamespacePicker.vue';
import UserMenu from './UserMenu.vue';

defineProps<{ product: Product; showNamespacePicker: boolean }>();

const ui = useUiStore();
const { t } = useI18n();
</script>

<template>
  <header class="topbar">
    <button
      type="button"
      class="hamburger"
      :aria-label="t('products.switcher')"
      @click="ui.openProductSwitcher"
    >
      <Menu :size="18" aria-hidden="true" />
    </button>

    <RouterLink :to="product.defaultRoute" class="context">
      <ProductBadge :icon="product.icon" :color="product.badgeColor" size="sm" />
      <span class="product-label">{{ t(product.labelKey) }}</span>
    </RouterLink>

    <div class="spacer" />

    <NamespacePicker v-if="showNamespacePicker" />

    <UserMenu />
  </header>
</template>

<style scoped>
.topbar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.45rem 1.25rem;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
  position: sticky;
  top: 0;
  z-index: 10;
  min-height: var(--topbar-height);
}

.hamburger {
  width: 2.25rem;
  height: 2.25rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: inherit;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}

.hamburger:hover {
  background: var(--color-surface-alt);
}

.context {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: inherit;
  text-decoration: none;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
}

.context:hover {
  background: var(--color-surface-alt);
}

.product-label {
  font-weight: 600;
  font-size: 0.95rem;
}

.spacer {
  flex: 1;
}
</style>
