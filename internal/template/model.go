package template

//const RegisterModel = NotEditMark + `
//package {{.StructInfo.Package}}
//
//func RegisterDo(do gen.DO) {
//	switch do.TableName() {
//	case "cash":
//		cashDo = do
//	}
//}
//
//`

// Model used as a variable because it cannot load template file after packed, params still can pass file
const Model = NotEditMark + `
package {{.Package}}

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	{{range .ImportPkgPaths}}{{.}} ` + "\n" + `{{end}}
)
{{if .TableName -}}const TableName{{.ModelStructName}} = "{{.TableName}}"{{- end}}

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


{{range .Fields -}}
	{{if .IsRelation -}}
func(r *{{$.ModelStructName}}) Query{{.Relation.Name}}() I{{.Relation.Name}}Do {
	return Query{{.Relation.Name}}.Key(r.{{.Relation.Key}})
}

func(r *{{$.ModelStructName}}) Get{{.Relation.Name}}(update ...bool) ({{if eq .Relation.Relationship "HasMany" "ManyToMany"}}[]{{end}}*{{.Relation.Type}}, error) {
	if len(update) == 0 && r.{{.Relation.Name}} != nil {
		return r.{{.Relation.Name}}, nil
	} 
	var data,err =  Query{{.Relation.Name}}.Key(r.{{.Relation.Key}}).First()
	r.{{.Relation.Name}} = data
	return r.{{.Relation.Name}},err
}

func(r *{{$.ModelStructName}}) Get{{.Relation.Name}}X(update ...bool) (*{{.Relation.Type}})  {
	var data,err = r.Get{{.Relation.Name}}(update...) 
	if err != nil {
		panic(err)
	}
	return data
}
{{end}}
{{end}}



`

// ModelMethod model struct DIY method
const ModelMethod = `

{{if .Doc -}}// {{.DocComment -}}{{end}}
func ({{.GetBaseStructTmpl}}){{.MethodName}}({{.GetParamInTmpl}})({{.GetResultParamInTmpl}}){{.Body}}
`
