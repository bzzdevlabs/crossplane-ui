<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { ApiError } from '@/services/api';
import { deleteGroup, listGroups, type GroupCR } from '@/services/users';

const { t } = useI18n();
const router = useRouter();

const loading = ref(true);
const error = ref<string | null>(null);
const items = ref<readonly GroupCR[]>([]);

async function refresh() {
  loading.value = true;
  error.value = null;
  try {
    const list = await listGroups();
    items.value = list.items ?? [];
  } catch (e) {
    error.value = e instanceof ApiError ? `${e.code} (${e.status})` : String(e);
  } finally {
    loading.value = false;
  }
}

async function remove(g: GroupCR) {
  if (!window.confirm(t('groups.confirmDelete', { name: g.metadata.name }))) return;
  try {
    await deleteGroup(g.metadata.name);
    await refresh();
  } catch (e) {
    error.value = e instanceof ApiError ? `${e.code} (${e.status})` : String(e);
  }
}

function open(g: GroupCR) {
  void router.push({ name: 'group-detail', params: { name: g.metadata.name } });
}

onMounted(() => {
  void refresh();
});
</script>

<template>
  <section class="groups">
    <header class="head">
      <div>
        <h1>{{ t('groups.title') }}</h1>
        <p class="muted">{{ t('groups.subtitle') }}</p>
      </div>
      <router-link class="btn primary" :to="{ name: 'group-create' }">
        + {{ t('groups.create') }}
      </router-link>
    </header>

    <p v-if="loading" class="muted">{{ t('common.loading') }}</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-else-if="items.length === 0" class="muted">{{ t('groups.empty') }}</p>

    <table v-else class="grid">
      <thead>
        <tr>
          <th>{{ t('columns.name') }}</th>
          <th>{{ t('groups.columns.displayName') }}</th>
          <th>{{ t('groups.columns.members') }}</th>
          <th aria-label="actions"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="g in items" :key="g.metadata.name" @click="open(g)">
          <td><code>{{ g.metadata.name }}</code></td>
          <td>{{ g.spec.displayName ?? '' }}</td>
          <td>{{ (g.status?.members ?? []).length }}</td>
          <td class="actions">
            <button type="button" class="btn link" @click.stop="remove(g)">
              {{ t('groups.delete') }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style scoped>
.groups {
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
