package templates

const Model = `//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package {{.Package}}{{if .HasImports}}

import ({{range .Imports}}
    "{{.}}"{{end}}
	{{- if .WithORM}}
	"context"
	{{- end}}
	"github.com/uptrace/bun"
){{end}}


{{range $model := .Entities}}
type {{.GoName}} struct {
	bun.BaseModel {{.Tag}}

	{{range .Columns}}
	{{.GoName}} {{.Type}} {{.Tag}} {{.Comment}}{{end}}{{if .HasRelations}}
	{{range .Relations}}
	{{.GoName}} *{{.GoType}} {{.Tag}} {{.Comment}}{{end}}{{end}}
}
{{end}}

/* Common ORM queries */

{{- if .WithORM}}
{{$dbstruct := .}}
/* 'SELECT' queries */
{{- range $model := .Entities}}
func (dbConn *{{ $dbstruct.ORMDbStruct }}) Select{{ .GoName }}() ([]*{{ .GoName }}, error) {
	ctx := context.Background()
	model := []*{{ .GoName }}{}
	{{$parent := .}}
	err := dbConn.NewSelect().
		{{- range .Columns}}
		{{- if $parent.NoAlias }}
		Column("{{ .Column.PGName -}}").
		{{- else}}
		Column("{{ $parent.Alias -}}.{{ .Column.PGName -}}").
		{{- end}}
		{{- end}}
		Model(&model).
		Scan(ctx)
	return model, err
}
{{end}}
{{- end}}

{{- if .WithSearch}}
/* Search Queries */

{{range $model := .Entities}}
type {{.GoName}}Search struct {
	search

	{{range .Columns}}
	{{.GoName}} {{.SearchType}}{{if .HasTags}} {{.Tag}}{{end}}{{end}}
}

func (s *{{.GoName}}Search) Apply(query bun.QueryBuilder) bun.QueryBuilder { {{range .Columns}}{{if .Relaxed}}
	if !reflect.ValueOf(s.{{.GoName}}).IsNil(){ {{else}}
	if s.{{.GoName}} != nil { {{end}}{{if .UseCustomRender}}
		{{.CustomRender}}{{else}}
		s.where(query, {{$model.GoName}}T.Table.Name(), Columns.{{$model.GoName}}.{{.GoName}}, s.{{.GoName}}){{end}}
	}{{end}}

	s.apply(query)

	return query
}

func (s *{{.GoName}}Search) Q() applier {
	return func(query bun.QueryBuilder) (bun.QueryBuilder, error) {
		return s.Apply(query), nil
	}
}
{{end}}
{{- end}}
`
