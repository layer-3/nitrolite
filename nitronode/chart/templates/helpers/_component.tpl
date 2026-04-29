{{/* vim: set filetype=mustache: */}}

{{/*
Returns Prometheus metrics' annotations depending on input Values
*/}}
{{- define "clearnode.component.metricsAnnotations" -}}
prometheus.io/scrape: {{ default false .enabled | print | quote }}
prometheus.io/port: {{ default "4242" .port | print | quote }}
prometheus.io/path: {{ default "/metrics" .endpoint | print | quote }}
{{- end }}

{{/*
Returns replica count depending on component and HPA settings
*/}}
{{- define "clearnode.component.replicaCount" -}}
{{- if not (and .autoscaling .autoscaling.enabled) }}
replicas: {{ .replicaCount }}
{{- end }}
{{- end }}

{{/*
Returns full docker image name 
*/}}
{{- define "clearnode.component.image" -}}
{{ printf "%s:%s" (print .repository) (print .tag) }}
{{- end }}

{{/*
Returns container ports configuration depending on input service
*/}}
{{- define "clearnode.component.ports" -}}
{{- if .http.enabled }}
ports:
{{- with .http }}
- name: http
  containerPort: {{ default .port .internalPort }}
  protocol: TCP
{{- end }}
{{- end }}
{{- end }}

{{/*
Returns component's resource consumption
*/}}
{{- define "clearnode.component.resources" -}}
resources:
  requests:
    cpu: {{ default "100m" .requests.cpu }}
    memory: {{ default "128Mi" .requests.memory }}
    ephemeral-storage: {{ default "100Mi" .requests.memory }}
  limits:
    cpu: {{ default "100m" .requests.cpu }}
    memory: {{ default "128Mi" .requests.memory }}
    ephemeral-storage: {{ default "100Mi" .requests.memory }}
{{- end }}

{{/*
Returns component's probes
*/}}
{{- define "clearnode.component.probes" -}}
{{- $port := default .Values.service.http.port .Values.service.http.internalPort }}
{{- range $name, $probe := .Values.probes }}
{{- if $probe.enabled }}
{{ printf "%sProbe" $name }}:
  {{- if eq $probe.type "http" }}
  httpGet:
    port: {{ $port }}
    path: {{ default "/health" $probe.endpoint }}
  {{- else }}
  tcpSocket:
    port: {{ $port }}
  {{- end }}
  initialDelaySeconds: {{ default 5 $probe.initialDelaySeconds }}
  timeoutSeconds: {{ default 10 $probe.timeoutSeconds }}
  periodSeconds: {{ default 10 $probe.periodSeconds }}
{{- end }}
{{- end }}
{{- end }}
