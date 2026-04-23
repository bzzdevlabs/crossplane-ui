<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { ApiError, deleteConnector, listConnectors, type ConnectorCR } from '@/services/api';

const { t } = useI18n();
const router = useRouter();

const loading = ref(true);
const error = ref<string | null>(null);
const items = ref<readonly ConnectorCR[]>([]);

async function refresh() {
  loading.value = true;
  error.value = null;
  try {
    const list = await listConnectors();
    items.value = list.items ?? [];
  } catch (e) {
    error.value = e instanceof ApiError ? `${e.code} (${e.status})` : String(e);
  } finally {
    loading.value = false;
  }
}

function readyCondition(c: ConnectorCR): string {
  const cond = c.status?.conditions?.find((x) => x.type === 'Ready');
  if (!cond) return '—';
  return cond.status === 'True' ? '✓' : cond.reason ?? cond.status;
}

async function remove(c: ConnectorCR) {
  if (!window.confirm(t('connectors.confirmDelete', { name: c.metadata.name }))) {
    return;
  }
  try {
    await deleteConnector(c.metadata.name);
    await refresh();
  } catch (e) {
    error.value = e instanceof ApiError ? `${e.code} (${e.status})` : String(e);
  }
}

function open(c: ConnectorCR) {
  void router.push({ name: 'connector-detail', params: { id: c.metadata.name } });
}

onMounted(() => {
  void refresh();
});
</script>

<template>
  <section class="connectors">
    <header class="head">
      <div>
        <h1>{{ t('connectors.title') }}</h1>
        <p class="muted">{{ t('connectors.subtitle') }}</p>
      </div>
      <router-link class="btn primary" :to="{ name: 'connector-create' }">
        + {{ t('connectors.create') }}
      </router-link>
    </header>

    <p v-if="loading" class="muted">{{ t('common.loading') }}</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-else-if="items.length === 0" class="muted">{{ t('connectors.empty') }}</p>

    <table v-else class="grid">
      <thead>
        <tr>
          <th>{{ t('columns.name') }}</th>
          <th>{{ t('connectors.columns.type') }}</th>
          <th>{{ t('connectors.columns.id') }}</th>
          <th>{{ t('columns.status') }}</th>
          <th>{{ t('connectors.columns.disabled') }}</th>
          <th aria-label="actions"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="c in items" :key="c.metadata.name" @click="open(c)">
          <td>{{ c.spec.name }}</td>
          <td><code>{{ c.spec.type }}</code></td>
          <td><code>{{ c.spec.id }}</code></td>
          <td>{{ readyCondition(c) }}</td>
          <td>{{ c.spec.disabled ? '✓' : '' }}</td>
          <td class="actions">
            <button type="button" class="btn link" @click.stop="remove(c)">
              {{ t('connectors.delete') }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style scoped>
.connectors {
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
  text-decoration: none;
  cursor: pointer;
  background: transparent;
  color: inherit;
  font: inherit;
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
