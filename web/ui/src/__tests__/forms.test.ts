import { describe, expect, it } from 'vitest';

import { getPath, setPath } from '@/components/forms/path';
import { FORM_SCHEMAS, schemaForObject } from '@/components/forms/schemas';

describe('path helpers', () => {
  it('creates intermediate maps on set', () => {
    const obj: Record<string, unknown> = {};
    setPath(obj, ['spec', 'credentials', 'source'], 'Secret');
    expect(getPath(obj, ['spec', 'credentials', 'source'])).toBe('Secret');
  });

  it('replaces arrays with fresh maps when walking through them', () => {
    const obj: Record<string, unknown> = { spec: { versions: [{ name: 'v1' }] } };
    setPath(obj, ['spec', 'versions', 'name'], 'v2');
    expect((obj.spec as { versions: { name: string } }).versions.name).toBe('v2');
  });

  it('deletes the leaf when setting empty or undefined', () => {
    const obj: Record<string, unknown> = { metadata: { name: 'x' } };
    setPath(obj, ['metadata', 'name'], '');
    expect((obj.metadata as Record<string, unknown>).name).toBeUndefined();
  });
});

describe('form schema resolution', () => {
  it('maps each known apiVersion/kind to its schema', () => {
    const ids = FORM_SCHEMAS.map((s) => s.id);
    expect(ids).toEqual(['composition', 'xrd', 'provider', 'function', 'providerconfig']);
  });

  it('matches a Composition object to the composition schema', () => {
    const schema = schemaForObject({
      apiVersion: 'apiextensions.crossplane.io/v1',
      kind: 'Composition',
    });
    expect(schema?.id).toBe('composition');
  });

  it('matches arbitrary ProviderConfig kinds', () => {
    const schema = schemaForObject({
      apiVersion: 'aws.upbound.io/v1beta1',
      kind: 'ProviderConfig',
    });
    expect(schema?.id).toBe('providerconfig');
  });

  it('returns null for unknown kinds', () => {
    expect(schemaForObject({ apiVersion: 'v1', kind: 'Pod' })).toBeNull();
  });
});
