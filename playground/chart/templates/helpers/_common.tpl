{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the component.
*/}}
{{- define "playground.common.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully common name.
If release name contains chart name it will be used as a full name.
*/}}
{{- define "playground.common.fullname" -}}
{{- if .Values.prefixOverride }}
{{- printf "%s-%s" .Values.prefixOverride .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- if contains .Chart.Name .Release.Name }}
{{- print .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Common Selector labels
*/}}
{{- define "playground.common.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "playground.common.labels" -}}
helm.sh/chart: {{ include "playground.common.chart" . }}
{{ include "playground.common.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.extraLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "playground.common.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Returns common image pull secrets
*/}}
{{- define "playground.common.imagePullSecrets" -}}
{{- with .Values.imagePullSecret }}
imagePullSecrets:
- name: {{ . }}
{{- end }}
{{- end }}

{{/*
Returns common environment variables. The nginx container itself needs no
configuration — the entrypoint reads these values to render
/v1/playground/env.js before nginx starts.
*/}}
{{- define "playground.common.env" -}}
- name: NITRONODE_URL
  value: {{ .Values.config.nitronodeWsUrl | quote }}
- name: FAUCET_ENABLED
  value: {{ .Values.config.faucetEnabled | print | quote }}
{{- end }}

{{/*
Returns common node selector labels
*/}}
{{- define "playground.common.nodeSelectorLabels" -}}
{{- with .Values.nodeSelector }}
nodeSelector:
  {{ toYaml . | nindent 2 }}
{{- end }}
{{- end }}

{{/*
Returns common tolerations
*/}}
{{- define "playground.common.tolerations" -}}
{{- with .Values.tolerations }}
tolerations:
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Returns common pod's affinity
*/}}
{{- define "playground.common.affinity" -}}
{{- with .Values.affinity }}
affinity:
  {{ toYaml . | nindent 2 }}
{{- end }}
{{- end }}
