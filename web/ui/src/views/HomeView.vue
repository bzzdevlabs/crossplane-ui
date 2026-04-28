<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import ProductBadge from '@/components/ui/ProductBadge.vue';
import { productById } from '@/products';

const { t } = useI18n();

const home = productById('home')!;
</script>

<template>
  <section class="placeholder">
    <ProductBadge :icon="home.icon" :color="home.badgeColor" size="lg" />
    <h1>{{ t(home.labelKey) }}</h1>
    <p>{{ t(home.subtitleKey ?? 'products.home.subtitle') }}</p>
  </section>
</template>

<style scoped>
.placeholder {
  min-height: 60vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: 0.75rem;
}

h1 {
  margin: 0;
  font-size: 1.35rem;
  font-weight: 600;
}

p {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 0.9rem;
  max-width: 28rem;
}
</style>
