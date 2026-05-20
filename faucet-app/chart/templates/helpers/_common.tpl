{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the component.
*/}}
{{- define "faucet-app.common.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully common name.
If release name contains chart name it will be used as a full name.
*/}}
{{- define "faucet-app.common.fullname" -}}
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
{{- define "faucet-app.common.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "faucet-app.common.labels" -}}
helm.sh/chart: {{ include "faucet-app.common.chart" . }}
{{ include "faucet-app.common.selectorLabels" . }}
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
{{- define "faucet-app.common.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Returns common image pull secrets
*/}}
{{- define "faucet-app.common.imagePullSecrets" -}}
{{- with .Values.imagePullSecret }}
imagePullSecrets:   
- name: {{ . }}
{{- end }}
{{- end }}

{{/*
Returns common environment variables
*/}}
{{- define "faucet-app.common.env" -}}
- name: LOG_LEVEL
  value: {{ .Values.config.logLevel }}
- name: SERVER_PORT
  value: {{ .Values.service.http.internalPort | default "8080" | print | quote }}
- name: CLEARNODE_URL
  value: {{ .Values.config.clearnodeWsUrl | print }}
- name: TOKEN_SYMBOL
  value: {{ .Values.config.token.symbol | print }}
- name: STANDARD_TIP_AMOUNT
  value: {{ .Values.config.token.tipAmount | print | quote }}
- name: MIN_TRANSFER_COUNT
  value: {{ .Values.config.minTransferCount | print | quote }}
{{- range $key, $value := .Values.config.extraEnvs }}
- name: {{ $key | upper }}
  value: {{ $value | print | quote }}
{{- end }}
{{- end }}

{{/*
Returns common node selector labels
*/}}
{{- define "faucet-app.common.nodeSelectorLabels" -}}
{{- with .Values.nodeSelector }}
nodeSelector:
  {{ toYaml . | nindent 2 }}
{{- end }}
{{- end }}

{{/*
Returns common tolerations
*/}}
{{- define "faucet-app.common.tolerations" -}}
{{- with .Values.tolerations }}
tolerations:
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Returns common pod's affinity
*/}}
{{- define "faucet-app.common.affinity" -}}
{{- with .Values.affinity }}
affinity:
  {{ toYaml . | nindent 2 }}
{{- end }}
{{- end }}
