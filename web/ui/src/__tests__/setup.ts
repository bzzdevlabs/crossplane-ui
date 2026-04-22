/* Shared setup for Vitest.
 *
 * jsdom opaque-origin quirks leave window.localStorage with a malformed backing
 * store ("this._store.getItem is not a function" when other code, like
 * oidc-client-ts, pokes at it). We substitute a minimal in-memory polyfill so
 * every suite runs against the same known-good Storage implementation.
 */

class MemoryStorage implements Storage {
  private data = new Map<string, string>();

  get length(): number {
    return this.data.size;
  }
  key(index: number): string | null {
    return Array.from(this.data.keys())[index] ?? null;
  }
  getItem(key: string): string | null {
    return this.data.get(key) ?? null;
  }
  setItem(key: string, value: string): void {
    this.data.set(key, String(value));
  }
  removeItem(key: string): void {
    this.data.delete(key);
  }
  clear(): void {
    this.data.clear();
  }
}

function install(name: 'localStorage' | 'sessionStorage'): void {
  Object.defineProperty(window, name, {
    configurable: true,
    value: new MemoryStorage(),
  });
}

install('localStorage');
install('sessionStorage');
