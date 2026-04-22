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

async function logout(): Promise<void> {
  await auth.signOut();
}
</script>

<template>
  <div class="layout">
    <aside class="sidebar" aria-label="Primary navigation">
      <div class="brand">
        <span class="brand-mark">cx</span>
        <div>
          <div class="brand-title">{{ t('app.title') }}</div>
          <div class="brand-subtitle">{{ t('app.subtitle') }}</div>
        </div>
      </div>

      <nav>
        <div class="nav-section">
          <div class="nav-heading">{{ t('nav.sections.cluster') }}</div>
          <ul>
            <li>
              <RouterLink :to="{ name: 'home' }">{{ t('nav.home') }}</RouterLink>
            </li>
          </ul>
        </div>

        <div class="nav-section">
          <div class="nav-heading">{{ t('nav.sections.crossplane') }}</div>
          <ul>
            <li><span class="disabled">{{ t('nav.compositions') }}</span></li>
            <li><span class="disabled">{{ t('nav.composites') }}</span></li>
            <li><span class="disabled">{{ t('nav.managed') }}</span></li>
            <li><span class="disabled">{{ t('nav.providers') }}</span></li>
            <li><span class="disabled">{{ t('nav.functions') }}</span></li>
          </ul>
        </div>

        <div class="nav-section">
          <div class="nav-heading">{{ t('nav.sections.administration') }}</div>
          <ul>
            <li><span class="disabled">{{ t('nav.users') }}</span></li>
            <li><span class="disabled">{{ t('nav.settings') }}</span></li>
          </ul>
        </div>
      </nav>
    </aside>

    <div class="main">
      <header class="topbar">
        <div class="spacer" />
        <button type="button" class="locale-toggle" @click="toggleLocale">
          {{ locale.toUpperCase() }}
        </button>
        <div class="user">
          <span v-if="auth.user" class="user-name">{{ auth.user.displayName }}</span>
          <button type="button" class="logout" @click="logout">{{ t('auth.logout') }}</button>
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
  background: var(--color-bg);
}

.sidebar {
  background: var(--color-surface-alt);
  border-right: 1px solid var(--color-border);
  padding: 1rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  position: sticky;
  top: 0;
  height: 100vh;
  overflow-y: auto;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0 0.5rem;
}

.brand-mark {
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  border-radius: 6px;
  background: var(--color-accent);
  color: var(--color-on-accent);
  font-weight: 700;
  align-items: center;
  justify-content: center;
}

.brand-title {
  font-weight: 600;
  font-size: 1rem;
}

.brand-subtitle {
  color: var(--color-text-muted);
  font-size: 0.75rem;
}

.nav-section + .nav-section {
  margin-top: 0.25rem;
}

.nav-heading {
  padding: 0 0.75rem;
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--color-text-muted);
  margin-bottom: 0.25rem;
}

.sidebar nav ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.sidebar nav a,
.sidebar nav .disabled {
  display: block;
  padding: 0.45rem 0.75rem;
  border-radius: 6px;
  color: inherit;
  text-decoration: none;
  font-size: 0.9rem;
}

.sidebar nav a:hover {
  background: var(--color-surface);
}

.sidebar nav a.router-link-exact-active {
  background: var(--color-accent-subtle);
  color: var(--color-accent);
  font-weight: 500;
}

.sidebar nav .disabled {
  color: var(--color-text-muted);
  cursor: not-allowed;
}

.main {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.topbar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 1.25rem;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
  position: sticky;
  top: 0;
  z-index: 1;
}

.spacer {
  flex: 1;
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

.user {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.user-name {
  color: var(--color-text-muted);
  font-size: 0.9rem;
}

.logout {
  padding: 0.25rem 0.6rem;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.content {
  flex: 1;
  padding: 1.5rem 1.75rem;
}
</style>
