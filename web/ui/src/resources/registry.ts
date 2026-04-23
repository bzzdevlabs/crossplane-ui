import { FORM_SCHEMAS, type FormSchema } from '@/components/forms/schemas';
import type { CrossplaneResource } from '@/services/api';

export type StatusVariant = 'ready' | 'degraded' | 'pending' | 'errored' | 'unknown';

// ResourceKind describes how a single Kubernetes/Crossplane kind is rendered
// in the list / detail / form templates. The `id` is the URL slug used in
// `/crossplane/:resource` and `category` maps onto the gateway's existing
// aggregated resource summary so list views can reuse that endpoint.
export interface ResourceKind {
  readonly id: string;
  readonly labelKey: string;
  readonly pluralLabelKey: string;
  readonly category?: 'composition' | 'composite' | 'managed' | 'provider' | 'function';
  readonly gvr?: { readonly group: string; readonly version: string; readonly resource: string };
  readonly form?: FormSchema;
  readonly matchKind?: (obj: { readonly apiVersion?: string; readonly kind?: string }) => boolean;
}

function findFormSchema(id: string): FormSchema | undefined {
  return FORM_SCHEMAS.find((s) => s.id === id);
}

export const RESOURCE_KINDS: readonly ResourceKind[] = [
  {
    id: 'compositions',
    labelKey: 'kinds.composition.label',
    pluralLabelKey: 'kinds.composition.plural',
    category: 'composition',
    gvr: { group: 'apiextensions.crossplane.io', version: 'v1', resource: 'compositions' },
    form: findFormSchema('composition'),
  },
  {
    id: 'xrds',
    labelKey: 'kinds.xrd.label',
    pluralLabelKey: 'kinds.xrd.plural',
    gvr: {
      group: 'apiextensions.crossplane.io',
      version: 'v2',
      resource: 'compositeresourcedefinitions',
    },
    form: findFormSchema('xrd'),
  },
  {
    id: 'composites',
    labelKey: 'kinds.composite.label',
    pluralLabelKey: 'kinds.composite.plural',
    category: 'composite',
  },
  {
    id: 'managed',
    labelKey: 'kinds.managed.label',
    pluralLabelKey: 'kinds.managed.plural',
    category: 'managed',
  },
  {
    id: 'providers',
    labelKey: 'kinds.provider.label',
    pluralLabelKey: 'kinds.provider.plural',
    category: 'provider',
    gvr: { group: 'pkg.crossplane.io', version: 'v1', resource: 'providers' },
    form: findFormSchema('provider'),
  },
  {
    id: 'functions',
    labelKey: 'kinds.function.label',
    pluralLabelKey: 'kinds.function.plural',
    category: 'function',
    gvr: { group: 'pkg.crossplane.io', version: 'v1', resource: 'functions' },
    form: findFormSchema('function'),
  },
  {
    id: 'providerconfigs',
    labelKey: 'kinds.providerconfig.label',
    pluralLabelKey: 'kinds.providerconfig.plural',
    form: findFormSchema('providerconfig'),
  },
];

export function resourceKindById(id: string): ResourceKind | undefined {
  return RESOURCE_KINDS.find((k) => k.id === id);
}

export function statusFromConditions(r: CrossplaneResource): StatusVariant {
  if (r.ready === 'True' && r.synced === 'True') return 'ready';
  if (r.ready === 'False' || r.synced === 'False') return 'degraded';
  if (r.ready === 'Unknown' || r.synced === 'Unknown' || r.ready === '' || r.synced === '') {
    return 'pending';
  }
  return 'unknown';
}
