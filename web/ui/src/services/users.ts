import { apiFetch } from '@/services/api';

export interface UserCR {
  readonly apiVersion: string;
  readonly kind: string;
  readonly metadata: { readonly name: string; readonly creationTimestamp?: string };
  readonly spec: {
    readonly email: string;
    readonly username: string;
    readonly passwordSecretRef?: { readonly name: string };
    readonly groups?: readonly string[];
    readonly disabled?: boolean;
  };
  readonly status?: {
    readonly passwordHash?: string;
    readonly conditions?: readonly {
      readonly type: string;
      readonly status: string;
      readonly reason?: string;
      readonly message?: string;
    }[];
  };
}

export interface UserList {
  readonly items: readonly UserCR[];
}

export interface GroupCR {
  readonly apiVersion: string;
  readonly kind: string;
  readonly metadata: { readonly name: string; readonly creationTimestamp?: string };
  readonly spec: {
    readonly displayName?: string;
    readonly description?: string;
  };
  readonly status?: {
    readonly members?: readonly string[];
  };
}

export interface GroupList {
  readonly items: readonly GroupCR[];
}

export function listUsers(): Promise<UserList> {
  return apiFetch<UserList>('/api/v1/auth/users');
}

export function getUser(name: string): Promise<UserCR> {
  return apiFetch<UserCR>(`/api/v1/auth/users?name=${encodeURIComponent(name)}`);
}

export function applyUser(body: Record<string, unknown>): Promise<UserCR> {
  return apiFetch<UserCR>('/api/v1/auth/users', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export function deleteUser(name: string): Promise<void> {
  return apiFetch<void>(`/api/v1/auth/users?name=${encodeURIComponent(name)}`, {
    method: 'DELETE',
  });
}

export function listGroups(): Promise<GroupList> {
  return apiFetch<GroupList>('/api/v1/auth/groups');
}

export function getGroup(name: string): Promise<GroupCR> {
  return apiFetch<GroupCR>(`/api/v1/auth/groups?name=${encodeURIComponent(name)}`);
}

export function applyGroup(body: Record<string, unknown>): Promise<GroupCR> {
  return apiFetch<GroupCR>('/api/v1/auth/groups', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export function deleteGroup(name: string): Promise<void> {
  return apiFetch<void>(`/api/v1/auth/groups?name=${encodeURIComponent(name)}`, {
    method: 'DELETE',
  });
}

export interface UserPasswordWrite {
  readonly namespace: string;
  readonly secretName: string;
  readonly password: string;
}

export function writeUserPassword(body: UserPasswordWrite): Promise<{ keys: string[] }> {
  return apiFetch<{ keys: string[] }>('/api/v1/auth/user-password', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}
