<script setup lang="ts" generic="T extends { readonly id: string }">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { RouteLocationRaw } from 'vue-router';

export interface Column<Row> {
  readonly key: string;
  readonly labelKey: string;
  readonly width?: string;
  readonly get?: (row: Row) => unknown;
}

const props = defineProps<{
  rows: readonly T[];
  columns: readonly Column<T>[];
  emptyKey?: string;
  rowTo?: (row: T) => RouteLocationRaw | null;
}>();

const { t } = useI18n();

const hasRows = computed(() => props.rows.length > 0);
</script>

<template>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th
            v-for="col in columns"
            :key="col.key"
            :style="col.width ? { width: col.width } : undefined"
          >
            {{ t(col.labelKey) }}
          </th>
          <th class="row-action" />
        </tr>
      </thead>
      <tbody v-if="hasRows">
        <tr v-for="row in rows" :key="row.id">
          <td v-for="col in columns" :key="col.key">
            <RouterLink v-if="col.key === 'name' && rowTo && rowTo(row)" :to="rowTo(row)!">
              <slot :name="`cell-${col.key}`" :row="row" :value="col.get ? col.get(row) : undefined">
                {{ col.get ? col.get(row) : '' }}
              </slot>
            </RouterLink>
            <slot
              v-else
              :name="`cell-${col.key}`"
              :row="row"
              :value="col.get ? col.get(row) : undefined"
            >
              {{ col.get ? col.get(row) : '' }}
            </slot>
          </td>
          <td class="row-action">
            <slot name="row-actions" :row="row" />
          </td>
        </tr>
      </tbody>
      <tbody v-else>
        <tr>
          <td :colspan="columns.length + 1" class="empty">
            {{ t(emptyKey ?? 'common.noRows') }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.table-wrap {
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
  overflow: hidden;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  text-align: left;
  padding: 0.6rem 0.9rem;
  border-bottom: 1px solid var(--color-border);
  font-size: 0.9rem;
  vertical-align: middle;
}

tbody tr:last-child td {
  border-bottom: 0;
}

th {
  font-weight: 600;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  background: var(--color-surface-alt);
}

tbody tr:hover {
  background: var(--color-accent-subtle);
}

.empty {
  text-align: center;
  color: var(--color-text-muted);
  padding: 2rem 0.9rem;
}

.row-action {
  width: 2.5rem;
  text-align: right;
}

td a {
  color: var(--color-accent);
  text-decoration: none;
}

td a:hover {
  text-decoration: underline;
}
</style>
