export interface AuthConfig {
  readonly enabled: boolean;
  readonly issuerURL?: string;
  readonly clientID?: string;
  readonly scopes?: readonly string[];
}

export interface GatewayConfig {
  readonly auth: AuthConfig;
  readonly authNamespace?: string;
  readonly version?: string;
}

let cached: GatewayConfig | null = null;

/**
 * Loads the gateway's public bootstrap document once per page load and caches
 * it. Intentionally fetched before the OIDC client is built so the SPA can
 * configure itself against whichever Dex instance the chart has wired up.
 */
export async function loadGatewayConfig(): Promise<GatewayConfig> {
  if (cached) {
    return cached;
  }
  const response = await fetch('/api/v1/config', { credentials: 'same-origin' });
  if (!response.ok) {
    throw new Error(`gateway config: HTTP ${response.status}`);
  }
  const parsed = (await response.json()) as GatewayConfig;
  cached = parsed;
  return parsed;
}

/** Reset the in-memory cache. Used by tests. */
export function resetGatewayConfigCache(): void {
  cached = null;
}
