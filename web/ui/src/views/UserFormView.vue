<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { ApiError } from '@/services/api';
import { loadGatewayConfig } from '@/services/config';
import {
  applyUser,
  getUser,
  listGroups,
  writeUserPassword,
  type GroupCR,
  type UserCR,
} from '@/services/users';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const isEdit = computed(() => route.name === 'user-detail');
const existingName = computed(() =>
  typeof route.params.name === 'string' ? route.params.name : undefined,
);

const loading = ref(false);
const saving = ref(false);
const error = ref<string | null>(null);

const authNamespace = ref('crossplane-ui');
const allGroups = ref<readonly GroupCR[]>([]);

interface FormState {
  metadataName: string;
  email: string;
  username: string;
  password: string;
  groups: string[];
  disabled: boolean;
  existingSecretRef?: string;
}

const form = reactive<FormState>({
  metadataName: '',
  email: '',
  username: '',
  password: '',
  groups: [],
  disabled: false,
});

function hydrate(u: UserCR) {
  form.metadataName = u.metadata.name;
  form.email = u.spec.email;
  form.username = u.spec.username;
  form.groups = [...(u.spec.groups ?? [])];
  form.disabled = u.spec.disabled ?? false;
  form.existingSecretRef = u.spec.passwordSecretRef?.name;
}

async function loadExisting() {
  if (!existingName.value) return;
  loading.value = true;
  try {
    const u = await getUser(existingName.value);
    hydrate(u);
  } catch (e) {
    error.value = e instanceof ApiError ? `${e.code} (${e.status})` : String(e);
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  const cfg = await loadGatewayConfig();
  if (cfg.authNamespace) authNamespace.value = cfg.authNamespace;
  try {
    const list = await listGroups();
    allGroups.value = list.items ?? [];
  } catch {
    // groups may be empty; list error is not fatal for the form.
  }
  if (isEdit.value) await loadExisting();
});

function toggleGroup(name: string) {
  const i = form.groups.indexOf(name);
  if (i >= 0) form.groups.splice(i, 1);
  else form.groups.push(name);
}

function sanitizeName(s: string): string {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9._-]/g, '-')
    .replace(/^-+|-+$/g, '');
}

async function save() {
  saving.value = true;
  error.value = null;
  try {
    const name = isEdit.value ? form.metadataName : sanitizeName(form.username);
    const secretName = form.existingSecretRef ?? `user-${name}`;

    if (!isEdit.value && !form.password) {
      throw new Error(t('users.errors.passwordRequired'));
    }
    if (form.password) {
      await writeUserPassword({
        namespace: authNamespace.value,
        secretName,
        password: form.password,
      });
    }

    const body = {
      apiVersion: 'auth.crossplane-ui.io/v1alpha1',
      kind: 'User',
      metadata: { name },
      spec: {
        email: form.email,
        username: form.username,
        groups: form.groups,
        disabled: form.disabled,
        passwordSecretRef: { name: secretName },
      },
    };
    await applyUser(body);
    void router.push({ name: 'users' });
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
        <h1>{{ isEdit ? t('users.edit') : t('users.create') }}</h1>
      </div>
      <router-link :to="{ name: 'users' }" class="btn">{{ t('common.back') }}</router-link>
    </header>

    <p v-if="loading" class="muted">{{ t('common.loading') }}</p>

    <div class="grid">
      <label>
        <span>{{ t('users.columns.username') }}<em class="req">*</em></span>
        <input
          v-model="form.username"
          :disabled="isEdit"
          required
          pattern="^[a-zA-Z0-9._-]+$"
          placeholder="alice"
        />
      </label>

      <label>
        <span>{{ t('users.columns.email') }}<em class="req">*</em></span>
        <input v-model="form.email" type="email" required placeholder="alice@example.com" />
      </label>

      <label>
        <span>
          {{ isEdit ? t('users.resetPassword') : t('users.password') }}
          <em v-if="!isEdit" class="req">*</em>
        </span>
        <input
          v-model="form.password"
          type="password"
          :placeholder="isEdit ? t('users.passwordKeepBlank') : ''"
          :required="!isEdit"
        />
      </label>

      <label class="row-checkbox">
        <input v-model="form.disabled" type="checkbox" />
        <span>{{ t('users.disabled') }}</span>
      </label>
    </div>

    <h2>{{ t('users.sections.groups') }}</h2>
    <p v-if="allGroups.length === 0" class="muted">{{ t('users.noGroups') }}</p>
    <div v-else class="chips">
      <label v-for="g in allGroups" :key="g.metadata.name" class="chip-input">
        <input
          type="checkbox"
          :checked="form.groups.includes(g.metadata.name)"
          @change="toggleGroup(g.metadata.name)"
        />
        <span>{{ g.spec.displayName ?? g.metadata.name }}</span>
      </label>
    </div>

    <p v-if="error" class="error">{{ error }}</p>

    <div class="actions">
      <button type="button" class="btn primary" :disabled="saving" @click="save">
        {{ saving ? t('common.loading') : t('users.save') }}
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

.error {
  color: var(--color-danger, #c0392b);
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr));
  gap: 0.75rem 1rem;
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

input[type='text'],
input[type='email'],
input[type='password'],
input:not([type]) {
  padding: 0.4rem 0.5rem;
  border: 1px solid var(--color-border, #ccd);
  border-radius: 4px;
  font: inherit;
}

.row-checkbox {
  flex-direction: row;
  align-items: center;
  gap: 0.5rem;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.chip-input {
  flex-direction: row;
  align-items: center;
  gap: 0.35rem;
  padding: 0.25rem 0.6rem;
  border: 1px solid var(--color-border, #ccd);
  border-radius: 999px;
  font-size: 0.85rem;
  cursor: pointer;
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
