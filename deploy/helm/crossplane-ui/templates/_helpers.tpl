{{/*
Expand the name of the chart.
*/}}
{{- define "crossplane-ui.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a fully-qualified app name, capped at 63 chars, used as the
base for most resource names.
*/}}
{{- define "crossplane-ui.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Chart label (name-version).
*/}}
{{- define "crossplane-ui.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels applied to every resource.
*/}}
{{- define "crossplane-ui.labels" -}}
helm.sh/chart: {{ include "crossplane-ui.chart" . }}
app.kubernetes.io/name: {{ include "crossplane-ui.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: crossplane-ui
{{- with .Values.global.labels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{/*
Selector labels for a given component. Pass the component name via a dict,
e.g. {{ include "crossplane-ui.selectorLabels" (dict "ctx" . "component" "gateway") }}
*/}}
{{- define "crossplane-ui.selectorLabels" -}}
app.kubernetes.io/name: {{ include "crossplane-ui.name" .ctx }}
app.kubernetes.io/instance: {{ .ctx.Release.Name }}
app.kubernetes.io/component: {{ .component }}
{{- end -}}

{{/*
Resolved image reference. Uses .Values.global.imageRegistry as the prefix if
the per-service image.repository does not already look like a fully-qualified
reference (no "/" in the value).
*/}}
{{- define "crossplane-ui.image" -}}
{{- $svc := .service -}}
{{- $global := .ctx.Values.global -}}
{{- $tag := $svc.tag | default .ctx.Chart.AppVersion -}}
{{- $repo := $svc.repository -}}
{{- if and $global.imageRegistry (not (contains "/" $repo)) -}}
{{- printf "%s/%s:%s" $global.imageRegistry $repo $tag -}}
{{- else -}}
{{- printf "%s:%s" $repo $tag -}}
{{- end -}}
{{- end -}}

{{/*
ServiceAccount name for a given component.
*/}}
{{- define "crossplane-ui.serviceAccountName" -}}
{{- printf "%s-%s" (include "crossplane-ui.fullname" .ctx) .component | trunc 63 | trimSuffix "-" -}}
{{- end -}}
