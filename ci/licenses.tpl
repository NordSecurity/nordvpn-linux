# Third party dependencies

## EPL-2.0 License
The following dependencies are licensed under the EPL-2.0 license.
{{ range . -}}
	{{ if (eq .LicenseName "EPL-2.0") }}
* [{{ .Name }}]({{ .LicenseURL }})
	{{ else -}}
	{{- end }}
{{- end }}

## BSD-3-Clause License
The following dependencies are licensed under the BSD-3-Clause license.
{{ range . -}}
	{{ if (eq .LicenseName "BSD-3-Clause") }}
* [{{ .Name }}]({{ .LicenseURL }})
	{{ else -}}
	{{- end }}
{{- end }}

## BSD-2-Clause License
The following dependencies are licensed under the BSD-2-Clause license.
{{ range . -}}
	{{ if (eq .LicenseName "BSD-2-Clause") }}
* [{{ .Name }}]({{ .LicenseURL }})
	{{ else -}}
	{{- end }}
{{- end }}

## Apache-2.0 License
The following dependencies are licensed under the Apache-2.0 license.
{{ range . -}}
	{{ if (eq .LicenseName "Apache-2.0") }}
* [{{ .Name }}]({{ .LicenseURL }})
	{{ else -}}
	{{- end }}
{{- end }}

## MIT License
The following dependencies are licensed under the MIT license.
{{ range . -}}
	{{ if (eq .LicenseName "MIT") }}
* [{{ .Name }}]({{ .LicenseURL }})
	{{ else -}}
	{{- end }}
{{- end }}
