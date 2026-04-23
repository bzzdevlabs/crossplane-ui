<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { X } from 'lucide-vue-next';

import { useUiStore } from '@/stores/ui';
import { PRODUCTS } from '@/products';

import ProductBadgeLink from './ProductBadgeLink.vue';

const ui = useUiStore();
const { t } = useI18n();

function close(): void {
  ui.closeProductSwitcher();
}
</script>

<template>
  <div v-if="ui.productSwitcherOpen" class="backdrop" @click="close">
    <aside class="panel" role="dialog" :aria-label="t('products.switcher')" @click.stop>
      <header>
        <h2>{{ t('products.switcher') }}</h2>
        <button type="button" class="close" :aria-label="t('common.closeMenu')" @click="close">
          <X :size="16" aria-hidden="true" />
        </button>
      </header>
      <ul>
        <li v-for="p in PRODUCTS" :key="p.id">
          <ProductBadgeLink :product="p" @click="close" />
        </li>
      </ul>
    </aside>
  </div>
</template>

<style scoped>
.backdrop {
  position: fixed;
  inset: 0;
  background: rgb(0 0 0 / 35%);
  z-index: 20;
  display: flex;
}

.panel {
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  width: 22rem;
  max-width: 85vw;
  height: 100vh;
  padding: 1rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  box-shadow: 2px 0 16px rgb(0 0 0 / 20%);
}

header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 0.5rem;
}

h2 {
  margin: 0;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--color-text-muted);
  font-weight: 600;
}

.close {
  appearance: none;
  background: transparent;
  border: 0;
  font-size: 1.4rem;
  line-height: 1;
  cursor: pointer;
  color: inherit;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
}

.close:hover {
  background: var(--color-surface-alt);
}

ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  overflow-y: auto;
}
</style>
