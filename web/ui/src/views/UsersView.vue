<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { ApiError } from '@/services/api';
import { deleteUser, listUsers, type UserCR } from '@/services/users';

const { t } = useI18n();
const router = useRouter();

const loading = ref(true);
const error = ref<string | null>(null);
const items = ref<readonly UserCR[]>([]);
const query = ref('');

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  if (!q) return items.value;
  return items.value.filter(
    (u) =>
      u.spec.username.toLowerCase().includes(q) ||
      u.spec.email.toLowerCase().includes(q) ||
      (u.spec.groups ?? []).some((g) => g.toLowerCase().includes(q)),
  );
});

async function refresh() {
  loading.value = true;
  error.value = null;
  try {
    const list = await listUsers();
    items.value = list.items ?? [];
  } catch (e) {
    error.value = e instanceof ApiError ? `${e.code} (${e.status})` : String(e);
  } finally {
    loading.value = false;
  }
}

async function remove(u: UserCR) {
  if (!window.confirm(t('users.confirmDelete', { name: u.spec.username }))) return;
  try {
    await deleteUser(u.metadata.name);
    await refresh();
  } catch (e) {
    error.value = e instanceof ApiError ? `${e.code} (${e.status})` : String(e);
  }
}

function readyCondition(u: UserCR): string {
  const cond = u.status?.conditions?.find((c) => c.type === 'Ready');
  if (!cond) return '—';
  return cond.status === 'True' ? '✓' : (cond.reason ?? cond.status);
}

function open(u: UserCR) {
  void router.push({ name: 'user-detail', params: { name: u.metadata.name } });
}

onMounted(() => {
  void refresh();
});
</script>

<template>
  <section class="users">
    <header class="head">
      <div>
        <h1>{{ t('users.title') }}</h1>
        <p class="muted">{{ t('users.subtitle') }}</p>
      </div>
      <router-link class="btn primary" :to="{ name: 'user-create' }">
        + {{ t('users.create') }}
      </router-link>
    </header>

    <input
      v-model="query"
      class="filter"
      type="search"
      :placeholder="t('common.filter')"
    />

    <p v-if="loading" class="muted">{{ t('common.loading') }}</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-else-if="filtered.length === 0" class="muted">{{ t('users.empty') }}</p>

    <table v-else class="grid">
      <thead>
        <tr>
          <th>{{ t('users.columns.username') }}</th>
          <th>{{ t('users.columns.email') }}</th>
          <th>{{ t('users.columns.groups') }}</th>
          <th>{{ t('columns.status') }}</th>
          <th>{{ t('users.columns.disabled') }}</th>
          <th aria-label="actions"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="u in filtered" :key="u.metadata.name" @click="open(u)">
          <td>{{ u.spec.username }}</td>
          <td>{{ u.spec.email }}</td>
          <td>
            <span v-for="g in u.spec.groups ?? []" :key="g" class="chip">{{ g }}</span>
          </td>
          <td>{{ readyCondition(u) }}</td>
          <td>{{ u.spec.disabled ? '✓' : '' }}</td>
          <td class="actions">
            <button type="button" class="btn link" @click.stop="remove(u)">
              {{ t('users.delete') }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style scoped>
.users {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

h1 {
  margin: 0;
  font-size: 1.5rem;
}

.muted {
  color: var(--color-text-muted);
  margin: 0;
}

.error {
  color: var(--color-danger, #c0392b);
}

.filter {
  max-width: 24rem;
  padding: 0.4rem 0.6rem;
  border: 1px solid var(--color-border, #ccd);
  border-radius: 4px;
  font: inherit;
}

.grid {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.grid th,
.grid td {
  text-align: left;
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid var(--color-border, #e6e6e6);
}

.grid tbody tr {
  cursor: pointer;
}

.grid tbody tr:hover {
  background: var(--color-hover, #f6f7fb);
}

.chip {
  display: inline-block;
  padding: 0.05rem 0.5rem;
  margin-right: 0.25rem;
  border-radius: 10px;
  background: var(--color-hover, #eef2ff);
  font-size: 0.75rem;
}

.actions {
  text-align: right;
}

.btn {
  display: inline-block;
  padding: 0.4rem 0.8rem;
  border-radius: 4px;
  border: 1px solid transparent;
  cursor: pointer;
  background: transparent;
  color: inherit;
  font: inherit;
  text-decoration: none;
}

.btn.primary {
  background: var(--color-accent, #2e4fd9);
  color: #fff;
}

.btn.link {
  padding: 0;
  color: var(--color-danger, #c0392b);
}
</style>
