import { computed, ref, watch } from 'vue';
import { defineStore } from 'pinia';

import { listNamespaces, type NamespaceSummary } from '@/services/api';

const STORAGE_SIDEBAR = 'cx-ui.sidebarCollapsed';
const STORAGE_NAMESPACE = 'cx-ui.namespace';

function readBool(key: string, fallback: boolean): boolean {
  try {
    const v = window.localStorage.getItem(key);
    if (v === null) return fallback;
    return v === '1' || v === 'true';
  } catch {
    return fallback;
  }
}

function readString(key: string): string | null {
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

function writeBool(key: string, v: boolean): void {
  try {
    window.localStorage.setItem(key, v ? '1' : '0');
  } catch {
    // opaque origin, ignore.
  }
}

function writeString(key: string, v: string | null): void {
  try {
    if (v === null) window.localStorage.removeItem(key);
    else window.localStorage.setItem(key, v);
  } catch {
    // opaque origin, ignore.
  }
}

export const useUiStore = defineStore('ui', () => {
  const sidebarCollapsed = ref<boolean>(readBool(STORAGE_SIDEBAR, false));
  const productSwitcherOpen = ref(false);
  const userMenuOpen = ref(false);

  // `null` means "All namespaces" (no filter). Otherwise a single namespace
  // name. Richer scope groups (Only User / Only System / Only Cluster) can
  // land later without changing the API surface.
  const namespace = ref<string | null>(readString(STORAGE_NAMESPACE));

  const namespaces = ref<readonly NamespaceSummary[]>([]);
  const namespacesLoading = ref(false);
  const namespacesError = ref<string | null>(null);
  const namespacesLoaded = ref(false);

  const isAllNamespaces = computed(() => namespace.value === null);

  function setSidebarCollapsed(v: boolean): void {
    sidebarCollapsed.value = v;
  }

  function toggleSidebar(): void {
    sidebarCollapsed.value = !sidebarCollapsed.value;
  }

  function setNamespace(ns: string | null): void {
    namespace.value = ns;
  }

  function openProductSwitcher(): void {
    productSwitcherOpen.value = true;
    userMenuOpen.value = false;
  }

  function closeProductSwitcher(): void {
    productSwitcherOpen.value = false;
  }

  function toggleUserMenu(): void {
    userMenuOpen.value = !userMenuOpen.value;
  }

  function closeUserMenu(): void {
    userMenuOpen.value = false;
  }

  async function loadNamespaces(force = false): Promise<void> {
    if (namespacesLoading.value) return;
    if (namespacesLoaded.value && !force) return;
    namespacesLoading.value = true;
    namespacesError.value = null;
    try {
      const res = await listNamespaces();
      namespaces.value = res.items;
      namespacesLoaded.value = true;
    } catch (err) {
      namespacesError.value = err instanceof Error ? err.message : String(err);
    } finally {
      namespacesLoading.value = false;
    }
  }

  watch(sidebarCollapsed, (v) => writeBool(STORAGE_SIDEBAR, v));
  watch(namespace, (v) => writeString(STORAGE_NAMESPACE, v));

  return {
    sidebarCollapsed,
    productSwitcherOpen,
    userMenuOpen,
    namespace,
    namespaces,
    namespacesLoading,
    namespacesError,
    namespacesLoaded,
    isAllNamespaces,
    setSidebarCollapsed,
    toggleSidebar,
    setNamespace,
    openProductSwitcher,
    closeProductSwitcher,
    toggleUserMenu,
    closeUserMenu,
    loadNamespaces,
  };
});
