package template

const Field = `package {{.Package}}

var (
	ALL = {{.Dao}}.Query{{.StructName}}.ALL
{{range .Fields -}}
	{{if not .IsRelation -}}
		{{- if .ColumnName -}}{{.Name}} = {{$.Dao}}.Query{{$.StructName}}.{{.Name}}{{- end -}}
	{{- else -}}
		{{.Relation.Name}} = {{$.Dao}}.Query{{$.StructName}}.{{.Relation.Name}}
{{end}}
{{end}}
)
`
