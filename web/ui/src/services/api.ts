import { useAuthStore } from '@/stores/auth';

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

/**
 * apiFetch wraps `fetch` with the common bits every authenticated call to the
 * gateway needs: Bearer token header, JSON response handling, and a typed
 * error for non-2xx responses.
 */
export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const auth = useAuthStore();
  const headers = new Headers(init.headers);
  if (!headers.has('Accept')) {
    headers.set('Accept', 'application/json');
  }
  if (auth.idToken && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${auth.idToken}`);
  }

  const response = await fetch(path, {
    ...init,
    headers,
    credentials: 'same-origin',
  });

  if (response.status === 401) {
    // Token expired or revoked — drop the session so the router forces a
    // re-login on the next navigation.
    auth.clear();
  }

  if (!response.ok) {
    let code = 'http_error';
    let message = `HTTP ${response.status}`;
    try {
      const body = (await response.json()) as { error?: string; message?: string };
      if (body.error) code = body.error;
      if (body.message) message = body.message;
    } catch {
      // non-JSON body, keep defaults.
    }
    throw new ApiError(response.status, code, message);
  }

  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export interface NamespaceSummary {
  readonly name: string;
  readonly phase: string;
  readonly creationTimestamp: string;
  readonly labels?: Readonly<Record<string, string>>;
}

export interface NamespacesResponse {
  readonly items: readonly NamespaceSummary[];
}

export function listNamespaces(): Promise<NamespacesResponse> {
  return apiFetch<NamespacesResponse>('/api/v1/namespaces');
}

export type ConditionStatus = string;

export interface CrossplaneResource {
  readonly apiVersion: string;
  readonly kind: string;
  readonly resource: string;
  readonly name: string;
  readonly namespace?: string;
  readonly ready: ConditionStatus;
  readonly synced: ConditionStatus;
  readonly creationTimestamp: string;
}

export interface CrossplaneGroup {
  readonly category: string;
  readonly items: readonly CrossplaneResource[];
  readonly error?: string;
}

export interface CrossplaneSummary {
  readonly groups: readonly CrossplaneGroup[];
}

export function listCrossplaneResources(): Promise<CrossplaneSummary> {
  return apiFetch<CrossplaneSummary>('/api/v1/crossplane/resources');
}

export interface ResourceRef {
  readonly group: string;
  readonly version: string;
  readonly resource: string;
  readonly namespace?: string;
  readonly name: string;
}

function resourceQuery(ref: ResourceRef, extra: Record<string, string> = {}): string {
  const params = new URLSearchParams({
    group: ref.group,
    version: ref.version,
    resource: ref.resource,
    name: ref.name,
    ...extra,
  });
  if (ref.namespace) {
    params.set('namespace', ref.namespace);
  }
  return params.toString();
}

export function getResource<T = unknown>(ref: ResourceRef): Promise<T> {
  return apiFetch<T>(`/api/v1/crossplane/resource?${resourceQuery(ref)}`);
}

export interface ApplyOptions {
  readonly dryRun?: boolean;
}

export function applyResource<T = unknown>(
  ref: ResourceRef,
  object: unknown,
  opts: ApplyOptions = {},
): Promise<T> {
  const extra: Record<string, string> = {};
  if (opts.dryRun) extra.dryRun = 'All';
  return apiFetch<T>(`/api/v1/crossplane/resource?${resourceQuery(ref, extra)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(object),
  });
}

export function deleteResource(ref: ResourceRef): Promise<void> {
  return apiFetch<void>(`/api/v1/crossplane/resource?${resourceQuery(ref)}`, {
    method: 'DELETE',
  });
}

export type ConnectorType = 'ldap' | 'saml' | 'github' | 'google' | 'oidc';

export interface ConnectorSecretRef {
  readonly name: string;
  readonly key: string;
}

export interface ConnectorSecretInjection {
  readonly path: string;
  readonly secretRef: ConnectorSecretRef;
}

export interface ConnectorSpec {
  readonly id: string;
  readonly type: ConnectorType;
  readonly name: string;
  readonly config: Record<string, unknown>;
  readonly secretRefs?: readonly ConnectorSecretInjection[];
  readonly disabled?: boolean;
}

export interface ConnectorCR {
  readonly apiVersion: string;
  readonly kind: string;
  readonly metadata: {
    readonly name: string;
    readonly creationTimestamp?: string;
    readonly resourceVersion?: string;
  };
  readonly spec: ConnectorSpec;
  readonly status?: {
    readonly conditions?: readonly {
      readonly type: string;
      readonly status: string;
      readonly reason?: string;
      readonly message?: string;
    }[];
  };
}

export interface ConnectorList {
  readonly items: readonly ConnectorCR[];
}

export function listConnectors(): Promise<ConnectorList> {
  return apiFetch<ConnectorList>('/api/v1/auth/connectors');
}

export function getConnector(name: string): Promise<ConnectorCR> {
  return apiFetch<ConnectorCR>(`/api/v1/auth/connectors?name=${encodeURIComponent(name)}`);
}

export function applyConnector(body: Record<string, unknown>): Promise<ConnectorCR> {
  return apiFetch<ConnectorCR>('/api/v1/auth/connectors', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export function deleteConnector(name: string): Promise<void> {
  return apiFetch<void>(`/api/v1/auth/connectors?name=${encodeURIComponent(name)}`, {
    method: 'DELETE',
  });
}

export interface ConnectorSecretWrite {
  readonly namespace: string;
  readonly name: string;
  readonly data: Record<string, string>;
}

export function writeConnectorSecret(body: ConnectorSecretWrite): Promise<{ keys: string[] }> {
  return apiFetch<{ keys: string[] }>('/api/v1/auth/connector-secrets', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

