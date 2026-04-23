import { markRaw, type Component } from 'vue';

import CompositionForm from './CompositionForm.vue';
import FunctionForm from './FunctionForm.vue';
import ProviderConfigForm from './ProviderConfigForm.vue';
import ProviderForm from './ProviderForm.vue';
import XRDForm from './XRDForm.vue';
import type { Obj } from './path';

// FormSchema is the static description of a Crossplane kind the UI can edit
// through a typed form. `ref` is the GVR used by the gateway's generic
// resource endpoint; `match` is the shape-inspection callback used to pick a
// schema when editing an existing object loaded from the cluster.
export interface FormSchema {
  readonly id: string;
  readonly label: string;
  readonly ref: { readonly group: string; readonly version: string; readonly resource: string };
  readonly skeleton: () => Obj;
  readonly component: Component;
  readonly match?: (obj: Obj) => boolean;
}

function skeletonComposition(): Obj {
  return {
    apiVersion: 'apiextensions.crossplane.io/v1',
    kind: 'Composition',
    metadata: { name: 'example-composition' },
    spec: {
      compositeTypeRef: { apiVersion: 'example.org/v1alpha1', kind: 'XExample' },
      mode: 'Pipeline',
      pipeline: [],
    },
  };
}

function skeletonXRD(): Obj {
  return {
    apiVersion: 'apiextensions.crossplane.io/v2',
    kind: 'CompositeResourceDefinition',
    metadata: { name: 'xexamples.example.org' },
    spec: {
      group: 'example.org',
      names: { kind: 'XExample', plural: 'xexamples' },
      scope: 'Namespaced',
      versions: [{ name: 'v1alpha1', served: true, referenceable: true }],
    },
  };
}

function skeletonProvider(): Obj {
  return {
    apiVersion: 'pkg.crossplane.io/v1',
    kind: 'Provider',
    metadata: { name: 'provider-kubernetes' },
    spec: { package: 'xpkg.upbound.io/crossplane-contrib/provider-kubernetes:v0.16.0' },
  };
}

function skeletonFunction(): Obj {
  return {
    apiVersion: 'pkg.crossplane.io/v1',
    kind: 'Function',
    metadata: { name: 'function-go-templating' },
    spec: {
      package: 'xpkg.upbound.io/crossplane-contrib/function-go-templating:v0.9.0',
    },
  };
}

function skeletonProviderConfig(): Obj {
  return {
    apiVersion: 'aws.upbound.io/v1beta1',
    kind: 'ProviderConfig',
    metadata: { name: 'default' },
    spec: { credentials: { source: 'Secret' } },
  };
}

function apiVersionIs(obj: Obj, wanted: string, kind: string): boolean {
  return obj.apiVersion === wanted && obj.kind === kind;
}

export const FORM_SCHEMAS: readonly FormSchema[] = [
  {
    id: 'composition',
    label: 'Composition',
    ref: { group: 'apiextensions.crossplane.io', version: 'v1', resource: 'compositions' },
    skeleton: skeletonComposition,
    component: markRaw(CompositionForm),
    match: (o) => apiVersionIs(o, 'apiextensions.crossplane.io/v1', 'Composition'),
  },
  {
    id: 'xrd',
    label: 'CompositeResourceDefinition',
    ref: {
      group: 'apiextensions.crossplane.io',
      version: 'v2',
      resource: 'compositeresourcedefinitions',
    },
    skeleton: skeletonXRD,
    component: markRaw(XRDForm),
    match: (o) =>
      o.kind === 'CompositeResourceDefinition' &&
      typeof o.apiVersion === 'string' &&
      (o.apiVersion === 'apiextensions.crossplane.io/v2' ||
        o.apiVersion === 'apiextensions.crossplane.io/v1'),
  },
  {
    id: 'provider',
    label: 'Provider',
    ref: { group: 'pkg.crossplane.io', version: 'v1', resource: 'providers' },
    skeleton: skeletonProvider,
    component: markRaw(ProviderForm),
    match: (o) => apiVersionIs(o, 'pkg.crossplane.io/v1', 'Provider'),
  },
  {
    id: 'function',
    label: 'Function',
    ref: { group: 'pkg.crossplane.io', version: 'v1', resource: 'functions' },
    skeleton: skeletonFunction,
    component: markRaw(FunctionForm),
    match: (o) => apiVersionIs(o, 'pkg.crossplane.io/v1', 'Function'),
  },
  {
    id: 'providerconfig',
    label: 'ProviderConfig',
    ref: { group: '', version: '', resource: '' },
    skeleton: skeletonProviderConfig,
    component: markRaw(ProviderConfigForm),
    match: (o) => o.kind === 'ProviderConfig',
  },
];

export function schemaForObject(obj: Obj): FormSchema | null {
  for (const s of FORM_SCHEMAS) {
    if (s.match && s.match(obj)) return s;
  }
  return null;
}
