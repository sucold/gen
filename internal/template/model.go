package template

const RegisterModel = NotEditMark + `
package {{.StructInfo.Package}}

func RegisterDo(do gen.DO) {
	switch do.TableName() {
	case "cash":
		cashDo = do
	}
}

`

// Model used as a variable because it cannot load template file after packed, params still can pass file
const Model = NotEditMark + `
package {{.StructInfo.Package}}

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	{{range .ImportPkgPaths}}{{.}} ` + "\n" + `{{end}}
)
{{if .TableName -}}const TableName{{.ModelStructName}} = "{{.TableName}}"{{- end}}

type _col_{{.TableName}} struct {
	ALL field.Asterisk
	{{range .Fields -}}
		{{if not .IsRelation -}}
			{{if .MultilineComment -}}
			/*
{{.ColumnComment}}
    		*/
			{{end -}}
			{{- if .ColumnName -}}{{.Name}} field.{{.GenType}}{{if not .MultilineComment}}{{if .ColumnComment}}// {{.ColumnComment}}{{end}}{{end}}{{- end -}}
		{{end}}
	{{end}}

	fieldMap  map[string]any
}

var (
	_{{.ModelStructName}} = gen.NewDo(db,&{{.ModelStructName}}{})
	_co_{{.TableName}} = _col_{{.TableName}}{
		ALL:field.NewAsterisk(TableName{{.ModelStructName}}),
		{{range .Fields -}}
		{{if not .IsRelation -}}{{- if .ColumnName -}}
	{{.Name}}:field.New{{.GenType}}(TableName{{$.ModelStructName}}, "{{.ColumnName}}"),{{- end -}}
		{{end}}
	{{end}}
	}

)




// {{.ModelStructName}} {{.StructComment}}
type {{.ModelStructName}} struct {
    {{range .Fields}}
	{{if .MultilineComment -}}
	/*
{{.ColumnComment}}
    */
	{{end -}}
    {{.Name}} {{.Type}} ` + "`{{.Tags}}` " +
	"{{if not .MultilineComment}}{{if .ColumnComment}}// {{.ColumnComment}}{{end}}{{end}}" +
	`{{end}}
}

func(r *{{.ModelStructName}}) Delete() (err error) {
	_,err = _{{.ModelStructName}}.Delete(r)
	return
}


`

// ModelMethod model struct DIY method
const ModelMethod = `

{{if .Doc -}}// {{.DocComment -}}{{end}}
func ({{.GetBaseStructTmpl}}){{.MethodName}}({{.GetParamInTmpl}})({{.GetResultParamInTmpl}}){{.Body}}
`
