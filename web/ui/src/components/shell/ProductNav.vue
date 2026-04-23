<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, type RouteLocationRaw } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { PanelLeftClose, PanelLeftOpen } from 'lucide-vue-next';

import ProductBadge from '@/components/ui/ProductBadge.vue';
import { useUiStore } from '@/stores/ui';
import type { Product } from '@/products';

const props = defineProps<{ product: Product }>();

const { t } = useI18n();
const route = useRoute();
const ui = useUiStore();

const currentName = computed(() => (route.name ? String(route.name) : ''));
const currentResource = computed(() => {
  const p = route.params as Record<string, string | string[] | undefined>;
  const v = p.resource;
  return typeof v === 'string' ? v : '';
});

function isActive(to: RouteLocationRaw, routeNames?: readonly string[]): boolean {
  if (typeof to === 'object' && to !== null && 'name' in to) {
    if (to.name === currentName.value) {
      const params = (to as { params?: Record<string, unknown> }).params;
      if (params && typeof params === 'object' && 'resource' in params) {
        return params.resource === currentResource.value;
      }
      return true;
    }
  }
  if (routeNames && routeNames.includes(currentName.value)) {
    if (typeof to === 'object' && to !== null && 'params' in to) {
      const params = (to as { params?: Record<string, unknown> }).params;
      if (params && typeof params === 'object' && 'resource' in params) {
        return params.resource === currentResource.value;
      }
    }
    return true;
  }
  return false;
}
</script>

<template>
  <aside class="sidebar" :class="{ collapsed: ui.sidebarCollapsed }" aria-label="Section">
    <header class="brand">
      <ProductBadge :icon="product.icon" :color="product.badgeColor" size="md" />
      <div v-if="!ui.sidebarCollapsed" class="brand-text">
        <div class="brand-title">{{ t(product.labelKey) }}</div>
        <div v-if="product.subtitleKey" class="brand-sub">{{ t(product.subtitleKey) }}</div>
      </div>
    </header>

    <nav>
      <section
        v-for="(group, gIdx) in props.product.groups"
        :key="gIdx"
        class="group"
      >
        <div v-if="group.labelKey && !ui.sidebarCollapsed" class="heading">
          {{ t(group.labelKey) }}
        </div>
        <ul>
          <li v-for="item in group.items" :key="t(item.labelKey)">
            <RouterLink
              :to="item.to"
              :class="{ 'item-active': isActive(item.to, item.routeNames) }"
              :title="ui.sidebarCollapsed ? t(item.labelKey) : undefined"
            >
              {{ ui.sidebarCollapsed ? t(item.labelKey).charAt(0) : t(item.labelKey) }}
            </RouterLink>
          </li>
        </ul>
      </section>
    </nav>

    <button
      type="button"
      class="collapse-toggle"
      :aria-label="ui.sidebarCollapsed ? t('common.openMenu') : t('common.closeMenu')"
      @click="ui.toggleSidebar"
    >
      <PanelLeftOpen v-if="ui.sidebarCollapsed" :size="16" aria-hidden="true" />
      <PanelLeftClose v-else :size="16" aria-hidden="true" />
    </button>
  </aside>
</template>

<style scoped>
.sidebar {
  background: var(--color-surface-alt);
  border-right: 1px solid var(--color-border);
  padding: 1rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  position: sticky;
  top: var(--topbar-height);
  height: calc(100vh - var(--topbar-height));
  overflow-y: auto;
  transition: width 0.15s ease;
}

.sidebar.collapsed {
  padding: 1rem 0.35rem;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0 0.25rem;
}

.brand-text {
  min-width: 0;
}

.brand-title {
  font-weight: 600;
  font-size: 0.95rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.brand-sub {
  color: var(--color-text-muted);
  font-size: 0.75rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

nav {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  flex: 1;
}

.group + .group {
  margin-top: 0.25rem;
}

.heading {
  padding: 0 0.75rem;
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--color-text-muted);
  margin-bottom: 0.25rem;
}

ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

a {
  display: block;
  padding: 0.45rem 0.75rem;
  border-radius: 6px;
  color: inherit;
  text-decoration: none;
  font-size: 0.9rem;
}

a:hover {
  background: var(--color-surface);
}

a.router-link-active,
a.item-active {
  background: var(--color-accent-subtle);
  color: var(--color-accent);
  font-weight: 500;
}

.collapse-toggle {
  margin-top: auto;
  align-self: flex-end;
  appearance: none;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-muted);
  width: 1.75rem;
  height: 1.75rem;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.collapse-toggle:hover {
  color: var(--color-text);
  background: var(--color-surface);
}
</style>
