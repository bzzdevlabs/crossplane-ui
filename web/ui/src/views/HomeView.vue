<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import { PRODUCTS } from '@/products';
import ProductBadgeLink from '@/components/shell/ProductBadgeLink.vue';

const { t } = useI18n();

const quickProducts = PRODUCTS.filter((p) => p.id !== 'home');
</script>

<template>
  <section class="home">
    <header class="hero">
      <h1>{{ t('home.welcome') }}</h1>
      <p>{{ t('home.welcomeHint') }}</p>
    </header>

    <h2>{{ t('home.quickActions') }}</h2>
    <ul class="products">
      <li v-for="p in quickProducts" :key="p.id">
        <ProductBadgeLink :product="p" />
      </li>
    </ul>
  </section>
</template>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  max-width: 64rem;
}

.hero {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

h1 {
  margin: 0;
  font-size: 1.75rem;
}

.hero p {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 0.95rem;
  max-width: 48rem;
}

h2 {
  margin: 0;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--color-text-muted);
  font-weight: 600;
}

.products {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(20rem, 1fr));
  gap: 0.5rem;
}
</style>
