{{/* vim: set filetype=mustache: */}}
{{/*
Returns HorizontalPodAutoscaler API version depending on K8s cluster version
*/}}
{{- define "clearnode.hpa.apiVersion" -}}
{{- if semverCompare ">=1.23-0" .Capabilities.KubeVersion.Version -}}
autoscaling/v2
{{- else -}}
autoscaling/v2beta2
{{- end }}
{{- end }}

{{/*
Returns HorizontalPodAutoscaler resource target utilization depending on K8s cluster version
*/}}
{{- define "clearnode.hpa.targetUtilization" -}}
{{- if semverCompare ">=1.23-0" .Capabilities.KubeVersion.Version -}}
target:
  type: Utilization
  averageUtilization: {{ .averageUtilization }}
{{- else -}}
targetAverageUtilization: {{ .averageUtilization }}
{{- end }}
{{- end }}
