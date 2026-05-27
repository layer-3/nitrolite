{{/* vim: set filetype=mustache: */}}

{{/*
Returns the HorizontalPodAutoscaler API version depending on the cluster.
autoscaling/v2 has been stable since K8s 1.23 — kept the fallback in case the
chart is rendered against an older bundled CRD.
*/}}
{{- define "playground.hpa.apiVersion" -}}
{{- if semverCompare ">=1.23-0" .Capabilities.KubeVersion.Version -}}
autoscaling/v2
{{- else -}}
autoscaling/v2beta2
{{- end }}
{{- end }}
