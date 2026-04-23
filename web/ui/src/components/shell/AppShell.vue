<script setup lang="ts">
import { computed } from 'vue';
import { RouterView, useRoute } from 'vue-router';

import { useUiStore } from '@/stores/ui';
import { productForRouteName } from '@/products';

import ProductNav from './ProductNav.vue';
import ProductSwitcher from './ProductSwitcher.vue';
import Topbar from './Topbar.vue';

const ui = useUiStore();
const route = useRoute();

const product = computed(() =>
  productForRouteName(route.name ? String(route.name) : null),
);

const showSidebar = computed(() => product.value.groups.length > 0);
const showNamespacePicker = computed(() => product.value.id === 'crossplane');
</script>

<template>
  <div class="shell" :class="{ 'has-sidebar': showSidebar, collapsed: ui.sidebarCollapsed }">
    <Topbar :product="product" :show-namespace-picker="showNamespacePicker" />
    <div class="body">
      <ProductNav v-if="showSidebar" :product="product" />
      <main class="content">
        <RouterView />
      </main>
    </div>
    <ProductSwitcher />
  </div>
</template>

<style scoped>
.shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--color-bg);
}

.body {
  display: grid;
  grid-template-columns: 1fr;
  flex: 1;
  min-height: 0;
}

.shell.has-sidebar .body {
  grid-template-columns: 240px 1fr;
}

.shell.has-sidebar.collapsed .body {
  grid-template-columns: 64px 1fr;
}

.content {
  flex: 1;
  padding: 1.5rem 1.75rem;
  min-width: 0;
}
</style>
