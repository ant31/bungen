package templates

const Tables = `//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package {{.Package}}{{if .HasImports}}

{{end}}

{{range .Entities}}
	type Columns{{.GoName}} struct{
		{{range $i, $e := .Columns}}{{if $i}}, {{end}}{{.GoName}}{{end}} string{{if .HasRelations}}
		{{range $i, $e := .Relations}}{{if $i}}, {{end}}{{.GoName}}{{end}} string{{end}}
	}
{{end}}

type ColumnsSt struct { {{range .Entities}}
	{{.GoName}} Columns{{.GoName}}{{end}}
}

var Columns = ColumnsSt{ {{range .Entities}}
	{{.GoName}}: Columns{{.GoName}}{ {{range .Columns}}
		{{.GoName}}: "{{.PGName}}",{{end}}{{if .HasRelations}}
		{{range .Relations}}
		{{.GoName}}: "{{.GoName}}",{{end}}{{end}}
	},
{{end}}
}

type TableInfo struct {
	name string
    alias string
}

func (t TableInfo) Name() string {
	return t.name
}

func (t TableInfo) Alias() string {
	return t.alias
}

{{range .Entities}}
type {{.GoName}}Table struct {
	Columns{{.GoName}}
	Table TableInfo
}

var {{.GoName}}T = {{.GoName}}Table {
	Table: TableInfo{name: "{{.PGFullName}}",{{if not .NoAlias}}alias: "{{.Alias}}",{{end}}},
	Columns{{.GoName}}: Columns.{{.GoName}},
}
{{end}}

type TablesSt struct { {{range .Entities}}
		{{.GoName}} {{.GoName}}Table{{end}}
}

var Tables = TablesSt { {{range .Entities}}
	{{.GoName}}: {{.GoName}}T,{{end}}
}

var T = Tables
`
