{{/*
Return the name of the chart with the release name
*/}}
{{- define "mongo.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Return the full name used throughout resources
*/}}
{{- define "mongo.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name (include "mongo.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}


