{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "chart.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "chart.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "labels" -}}
chart: {{ template "chart.chart" . }}
heritage: {{ .Release.Service }}
release: {{ .Release.Name }}
{{- end -}}

{{/*
Define checksums for configmaps and envs.
*/}}
{{- define "configmap-pv-blacklist-checksum" }}
{{- with .Values.diskBlackList }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{- define "configmap-pv-max-size-checksum" }}
{{- with .Values.diskMaxSizeLimit }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{- define "env-checksum" }}
{{- toYaml .Values.env | sha256sum }}
{{- end }}
