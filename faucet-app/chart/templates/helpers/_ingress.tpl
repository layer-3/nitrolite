{{/* vim: set filetype=mustache: */}}

{{/*
Returns Ingress API version depending on K8s cluster version
*/}}
{{- define "faucet-app.ingress.apiVersion" -}}
{{- if semverCompare ">=1.19-0" .Capabilities.KubeVersion.Version -}}
networking.k8s.io/v1
{{- else if semverCompare ">=1.14-0" .Capabilities.KubeVersion.Version -}}
networking.k8s.io/v1beta1
{{- else -}}
extensions/v1beta1
{{- end }}
{{- end }}

{{/*
Returns default Ingress annotations
*/}}
{{- define "faucet-app.ingress.annotations" -}}
kubernetes.io/ingress.class: {{ default "nginx" .Values.networking.ingress.className }}
{{- if .Values.networking.ingress.tls.enabled }}
kubernetes.io/tls-acme: "true"
cert-manager.io/cluster-issuer: {{ .Values.networking.tlsClusterIssuer }}
nginx.ingress.kubernetes.io/ssl-redirect: "true"
{{- end }}
{{- if .Values.networking.ingress.grpc }}
nginx.ingress.kubernetes.io/backend-protocol: "GRPC"
{{- end }}
{{- with .Values.networking.ingress.annotations }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Returns Ingress TLS configuration
*/}}
{{- define "faucet-app.ingress.tls" -}}
{{- if .Values.networking.ingress.tls.enabled }}
tls:
  - secretName: "{{ .Values.networking.externalHostname | replace "." "-" }}-tls"
    hosts:
      - "{{ .Values.networking.externalHostname }}"
{{- end }}
{{- end }}

{{/*
Returns Ingress host path configuration
*/}}
{{- define "faucet-app.ingress.httpPath" -}}
{{- $http := .Values.service.http }}
- path: {{ $http.path }}
  {{- if semverCompare ">=1.18-0" .Capabilities.KubeVersion.Version }}
  pathType: ImplementationSpecific
  {{- end }}
  backend:
    {{ $svcName := include "faucet-app.common.fullname" . }}
    {{ $svcPort := default $http.port $http.internalPort }}
    {{- if semverCompare ">=1.19-0" .Capabilities.KubeVersion.Version }}
    service:
      name: {{ $svcName }}
      port:
        number: {{ $svcPort }}
    {{- else }}
    serviceName: {{ $svcName }}
    servicePort: {{ $svcPort }}
    {{- end }}
{{- end }}
