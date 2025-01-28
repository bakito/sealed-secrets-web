{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "sealed-secrets-web.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "sealed-secrets-web.fullname" -}}
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
Create chart name and version as used by the chart label.
*/}}
{{- define "sealed-secrets-web.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "sealed-secrets-web.labels" -}}
app.kubernetes.io/name: {{ include "sealed-secrets-web.name" . }}
helm.sh/chart: {{ include "sealed-secrets-web.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ . | toYaml }}
{{- end }}
{{- end -}}

{{/*
Create the name of the service account to use.
*/}}
{{- define "sealed-secrets-web.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "sealed-secrets-web.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}


{{/*
Generate image args
*/}}
{{- define "sealed-secrets-web.imageArgs" -}}
{{- $args := list -}}
{{- if .Values.includeLocalNamespaceOnly }}
{{- $args = append $args (printf "--include-namespaces=%s" .Release.Namespace) }}
{{- end }}
{{- if .Values.sealedSecrets.certURL }}
  {{- $args = append $args (printf "--sealed-secrets-cert-url=%s" .Values.sealedSecrets.certURL ) }}
{{- else }}
  {{- if .Values.sealedSecrets.namespace }}
  {{- $args = append $args (printf "--sealed-secrets-service-namespace=%s" .Values.sealedSecrets.namespace ) }}
  {{- end }}
  {{- if .Values.sealedSecrets.serviceName  }}
  {{- $args = append $args (printf "--sealed-secrets-service-name=%s" .Values.sealedSecrets.serviceName) }}
  {{- end }}
{{- end }}
{{- if .Values.webContext }}
{{- $args = append $args (printf "--web-context=%s" .Values.webContext) }}
{{- end }}
{{- if .Values.initialSecretFile }}
{{- $args = append $args (printf "--initial-secret-file=%s" .Values.initialSecretFile) }}
{{- end }}
{{- if .Values.disableLoadSecrets  }}
{{- $args = append $args "--disable-load-secrets" }}
{{- end }}
{{- if .Values.webLogs  }}
{{- $args = append $args "--enable-web-logs" }}
{{- end }}

{{- toYaml $args }}
{{- end -}}
