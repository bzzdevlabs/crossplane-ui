<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { useAuthStore } from '@/stores/auth';
import { setLocale } from '@/i18n';

const auth = useAuthStore();
const { t, locale } = useI18n();

const open = ref(false);
const root = ref<HTMLElement | null>(null);

const initials = computed(() => {
  const n = auth.user?.displayName ?? auth.user?.username ?? '?';
  return n
    .split(/\s+|[-_@.]/)
    .map((p) => p.charAt(0).toUpperCase())
    .filter(Boolean)
    .slice(0, 2)
    .join('');
});

function toggle(): void {
  open.value = !open.value;
}

function close(): void {
  open.value = false;
}

function toggleLocale(): void {
  setLocale(locale.value === 'en' ? 'fr' : 'en');
}

async function logout(): Promise<void> {
  close();
  await auth.signOut();
}

function handleDocClick(e: MouseEvent): void {
  if (!root.value) return;
  if (!root.value.contains(e.target as Node)) close();
}

onMounted(() => document.addEventListener('click', handleDocClick));
onBeforeUnmount(() => document.removeEventListener('click', handleDocClick));
</script>

<template>
  <div ref="root" class="user">
    <button
      type="button"
      class="avatar"
      :aria-label="auth.user?.displayName ?? 'user'"
      :aria-expanded="open"
      @click="toggle"
    >
      <span>{{ initials || '?' }}</span>
    </button>

    <div v-if="open" class="menu" role="menu">
      <div class="identity">
        <div class="name">{{ auth.user?.displayName }}</div>
        <div v-if="auth.user?.email" class="email">{{ auth.user.email }}</div>
      </div>
      <hr />
      <button type="button" role="menuitem" class="item" @click="toggleLocale">
        {{ locale.toUpperCase() === 'EN' ? 'Français' : 'English' }}
      </button>
      <button type="button" role="menuitem" class="item danger" @click="logout">
        {{ t('auth.logout') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.user {
  position: relative;
}

.avatar {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 50%;
  border: 1px solid var(--color-border);
  background: var(--color-accent);
  color: var(--color-on-accent);
  font-weight: 600;
  font-size: 0.8rem;
  cursor: pointer;
  letter-spacing: 0.02em;
}

.menu {
  position: absolute;
  right: 0;
  top: calc(100% + 0.35rem);
  z-index: 15;
  width: 14rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  box-shadow: 0 6px 18px rgb(0 0 0 / 12%);
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.identity {
  padding: 0.35rem 0.5rem;
}

.name {
  font-weight: 600;
}

.email {
  color: var(--color-text-muted);
  font-size: 0.8rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

hr {
  border: 0;
  border-top: 1px solid var(--color-border);
  margin: 0.25rem 0;
}

.item {
  appearance: none;
  background: transparent;
  border: 0;
  color: inherit;
  font: inherit;
  padding: 0.4rem 0.5rem;
  border-radius: 4px;
  cursor: pointer;
  text-align: left;
}

.item:hover {
  background: var(--color-accent-subtle);
}

.item.danger {
  color: var(--color-danger);
}
</style>
