<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { ApiError } from '@/services/api';
import { applyGroup, getGroup, type GroupCR } from '@/services/users';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const isEdit = computed(() => route.name === 'group-detail');
const existingName = computed(() =>
  typeof route.params.name === 'string' ? route.params.name : undefined,
);

const loading = ref(false);
const saving = ref(false);
const error = ref<string | null>(null);

const form = reactive({
  name: '',
  displayName: '',
  description: '',
  members: [] as string[],
});

function hydrate(g: GroupCR) {
  form.name = g.metadata.name;
  form.displayName = g.spec.displayName ?? '';
  form.description = g.spec.description ?? '';
  form.members = [...(g.status?.members ?? [])];
}

async function loadExisting() {
  if (!existingName.value) return;
  loading.value = true;
  try {
    const g = await getGroup(existingName.value);
    hydrate(g);
  } catch (e) {
    error.value = e instanceof ApiError ? `${e.code} (${e.status})` : String(e);
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  if (isEdit.value) void loadExisting();
});

async function save() {
  saving.value = true;
  error.value = null;
  try {
    const body = {
      apiVersion: 'auth.crossplane-ui.io/v1alpha1',
      kind: 'Group',
      metadata: { name: form.name },
      spec: {
        displayName: form.displayName,
        description: form.description,
      },
    };
    await applyGroup(body);
    void router.push({ name: 'groups' });
  } catch (e) {
    error.value = e instanceof ApiError ? `${e.code} (${e.status})` : String(e);
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <section class="form">
    <header class="head">
      <div>
        <h1>{{ isEdit ? t('groups.edit') : t('groups.create') }}</h1>
      </div>
      <router-link :to="{ name: 'groups' }" class="btn">{{ t('common.back') }}</router-link>
    </header>

    <p v-if="loading" class="muted">{{ t('common.loading') }}</p>

    <div class="grid">
      <label>
        <span>{{ t('columns.name') }}<em class="req">*</em></span>
        <input
          v-model="form.name"
          :disabled="isEdit"
          required
          pattern="^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
          placeholder="crossplane-ui-admin"
        />
        <small class="help">{{ t('groups.help.name') }}</small>
      </label>

      <label>
        <span>{{ t('groups.columns.displayName') }}</span>
        <input v-model="form.displayName" placeholder="Administrators" />
      </label>

      <label class="wide">
        <span>{{ t('groups.columns.description') }}</span>
        <textarea v-model="form.description" rows="2"></textarea>
      </label>
    </div>

    <template v-if="isEdit">
      <h2>{{ t('groups.sections.members') }}</h2>
      <p v-if="form.members.length === 0" class="muted">{{ t('groups.noMembers') }}</p>
      <ul v-else class="members">
        <li v-for="m in form.members" :key="m">{{ m }}</li>
      </ul>
      <p class="muted small">{{ t('groups.membersHint') }}</p>
    </template>

    <p v-if="error" class="error">{{ error }}</p>

    <div class="actions">
      <button type="button" class="btn primary" :disabled="saving" @click="save">
        {{ saving ? t('common.loading') : t('groups.save') }}
      </button>
    </div>
  </section>
</template>

<style scoped>
.form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: 48rem;
}

.head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

h1 {
  margin: 0;
  font-size: 1.5rem;
}

h2 {
  margin: 1rem 0 0.5rem;
  font-size: 1.1rem;
}

.muted {
  color: var(--color-text-muted);
  margin: 0;
}

.small {
  font-size: 0.8rem;
}

.error {
  color: var(--color-danger, #c0392b);
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr));
  gap: 0.75rem 1rem;
}

.wide {
  grid-column: 1 / -1;
}

label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.9rem;
}

label span {
  font-weight: 500;
}

input,
textarea {
  padding: 0.4rem 0.5rem;
  border: 1px solid var(--color-border, #ccd);
  border-radius: 4px;
  font: inherit;
}

.help {
  color: var(--color-text-muted);
  font-size: 0.8rem;
}

.members {
  margin: 0;
  padding-left: 1.25rem;
}

.req {
  color: var(--color-danger, #c0392b);
  font-style: normal;
  margin-left: 0.2rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
}

.btn {
  display: inline-block;
  padding: 0.45rem 0.9rem;
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

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
