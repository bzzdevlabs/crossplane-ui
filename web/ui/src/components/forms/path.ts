import { computed, type WritableComputedRef } from 'vue';

// Nested path helpers for reactive Unstructured-style records. Forms write
// through `setPath` so intermediate maps are created on demand, and read
// through `getPath` so missing branches collapse to undefined.

export type Obj = Record<string, unknown>;

export function getPath(root: Obj, path: readonly string[]): unknown {
  let cur: unknown = root;
  for (const key of path) {
    if (cur == null || typeof cur !== 'object') return undefined;
    cur = (cur as Obj)[key];
  }
  return cur;
}

export function setPath(root: Obj, path: readonly string[], value: unknown): void {
  if (path.length === 0) return;
  let cur: Obj = root;
  for (let i = 0; i < path.length - 1; i++) {
    const key = path[i]!;
    const next = cur[key];
    if (next == null || typeof next !== 'object' || Array.isArray(next)) {
      const fresh: Obj = {};
      cur[key] = fresh;
      cur = fresh;
    } else {
      cur = next as Obj;
    }
  }
  const leaf = path[path.length - 1]!;
  if (value === undefined || value === '') {
    delete cur[leaf];
  } else {
    cur[leaf] = value;
  }
}

// useStringField binds a string-valued path on the form's reactive object to
// a writable computed. Assigning through the returned ref emits an
// `update:modelValue` on `emit` with a shallow clone carrying the change.
export function useStringField(
  get: () => Obj,
  emit: (v: Obj) => void,
  path: readonly string[],
  fallback = '',
): WritableComputedRef<string> {
  return computed({
    get: () => {
      const v = getPath(get(), path);
      return typeof v === 'string' ? v : fallback;
    },
    set: (value: string) => {
      const next = { ...get() };
      setPath(next, path, value);
      emit(next);
    },
  });
}
