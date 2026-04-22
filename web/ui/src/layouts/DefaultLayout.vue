<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router';
import { useI18n } from 'vue-i18n';

import { useAuthStore } from '@/stores/auth';
import { setLocale } from '@/i18n';

const { t, locale } = useI18n();
const auth = useAuthStore();

function toggleLocale(): void {
  setLocale(locale.value === 'en' ? 'fr' : 'en');
}
</script>

<template>
  <div class="layout">
    <aside class="sidebar" aria-label="Primary navigation">
      <div class="brand">
        <span class="brand-title">{{ t('app.title') }}</span>
        <span class="brand-subtitle">{{ t('app.subtitle') }}</span>
      </div>
      <nav>
        <ul>
          <li>
            <RouterLink :to="{ name: 'home' }">{{ t('nav.home') }}</RouterLink>
          </li>
          <!-- Populated in M5+ -->
        </ul>
      </nav>
    </aside>

    <div class="main">
      <header class="topbar">
        <button type="button" class="locale-toggle" @click="toggleLocale">
          {{ locale.toUpperCase() }}
        </button>
        <div class="user">
          <span v-if="auth.user">{{ auth.user.displayName }}</span>
          <button type="button" @click="auth.clear()">{{ t('auth.logout') }}</button>
        </div>
      </header>

      <main class="content">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped>
.layout {
  display: grid;
  grid-template-columns: 240px 1fr;
  min-height: 100vh;
}

.sidebar {
  background: var(--color-surface-alt);
  border-right: 1px solid var(--color-border);
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.brand {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.brand-title {
  font-weight: 600;
  font-size: 1.125rem;
}

.brand-subtitle {
  color: var(--color-text-muted);
  font-size: 0.8125rem;
}

.sidebar nav ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.sidebar nav a {
  display: block;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  color: inherit;
  text-decoration: none;
}

.sidebar nav a.router-link-active {
  background: var(--color-accent-subtle);
  color: var(--color-accent);
}

.main {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
}

.user {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.locale-toggle {
  font-family: inherit;
  font-size: 0.75rem;
  letter-spacing: 0.05em;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 0.25rem 0.5rem;
  cursor: pointer;
}

.content {
  flex: 1;
  padding: 1.5rem;
}
</style>
