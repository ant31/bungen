package templates

const ORM = `//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package {{.Package}}{{if .HasImports}}

import ({{range .Imports}}
    "{{.}}"{{end}}
	{{- if .WithORM}}
	"context"
	{{- end}}
	"github.com/uptrace/bun"
){{end}}


/* Common ORM queries */

// Just a wrapper around database connection
{{- if .WithORM}}
type {{ .ORMDbStruct }} struct {
	*bun.DB
}
{{- end}}
`
