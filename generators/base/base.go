package base

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"strings"

	bungen "github.com/ant31/bungen/lib"
	"github.com/ant31/bungen/model"
	"github.com/ant31/bungen/util"

	"github.com/spf13/cobra"
)

const (
	// Conn is connection string (-c) basic flag
	Conn = "conn"

	// Output is output filename (-o) basic flag
	Output = "output"

	// Tables is basic flag (-t) for tables to generate
	Tables = "tables"

	// FollowFKs is basic flag (-f) for generate foreign keys models for selected tables
	FollowFKs = "follow-fk"

	// Package for model files
	Pkg = "pkg"

	// uuid type flag
	uuidFlag = "uuid"

	// custom types flag
	customTypesFlag = "custom-types"

	// generate simple ORM queries
	withORM        = "with-orm"
	withValidation = "with-validation"
	withSearch     = "with-search"
	relaxed        = "search-relaxed"
	dbWrap         = "db-wrap"
	keepPK         = "keep-pk"
	noDiscard      = "no-discard"
	noAlias        = "no-alias"
	softDelete     = "soft-delete"
	json           = "json"
	jsonTag        = "json-tag"
)

// Gen is interface for all generators
type Gen interface {
	AddFlags(command *cobra.Command)
	ReadFlags(command *cobra.Command) error

	Generate() error
}

// Packer is a function that compile entities to package
type Packer func(entities []model.Entity) (interface{}, error)

// Options is common options for all generators
type Options struct {
	// URL connection string
	URL string

	// Output file path
	Output string

	// List of Tables to generate
	// Default []string{"public.*"}
	Tables []string

	// Generate model for foreign keys,
	// even if Tables not listed in Tables param
	// will not generate fks if schema not listed
	FollowFKs bool
	// Package sets package name for model
	// Works only with SchemaPackage = false
	Package string

	// Do not replace primary key name to ID
	KeepPK bool

	// Soft delete column
	SoftDelete string

	// use sql.Null... instead of pointers
	UseSQLNulls bool

	// Do not generate alias tag
	NoAlias bool

	// Do not generate discard_unknown_columns tag
	NoDiscard bool

	// Override type for json/jsonb
	JSONTypes map[string]string

	// Add json tag to models
	AddJSONTag bool

	// Generate basic ORM queries
	WithORM bool
	// Generate Search queries
	WithSearch bool
	// Generate Vallidation functions
	WithValidation bool
	// Strict types in filters
	Relaxed bool

	// Struct name for ORM queries. Works only when GenORM == true
	DBWrapName string
	// Custom types goes here
	CustomTypes model.CustomTypeMapping
}

// Def sets default options if empty
func (o *Options) Def() {
	if len(o.Tables) == 0 {
		o.Tables = []string{util.Join(util.PublicSchema, "*")}
	}

	if o.CustomTypes == nil {
		o.CustomTypes = model.CustomTypeMapping{}
	}
}

// Generator is base generator used in other generators
type Generator struct {
	bungen.Bungen
	Name string
}

// NewGenerator creates generator
func NewGenerator(url string, n string) Generator {
	return Generator{
		Bungen: bungen.New(url, nil),
		Name:   n,
	}
}

// AddFlags adds basic flags to command
func AddFlags(command *cobra.Command) {
	flags := command.Flags()

	flags.StringP(Conn, "c", "", "connection string to your postgres database")
	if err := command.MarkFlagRequired(Conn); err != nil {
		panic(err)
	}

	flags.StringP(Output, "o", "", "output file name")
	if err := command.MarkFlagRequired(Output); err != nil {
		panic(err)
	}

	flags.StringP(Pkg, "p", "", "package for model files. if not set last folder name in output path will be used")

	flags.StringSliceP(Tables, "t", []string{"public.*"}, "table names for model generation separated by comma\nuse 'schema_name.*' to generate model for every table in model")
	flags.BoolP(FollowFKs, "f", false, "generate models for foreign keys, even if it not listed in Tables\n")

	flags.Bool(uuidFlag, false, "use github.com/google/uuid as type for uuid")

	flags.StringSlice(customTypesFlag, []string{}, "set custom types separated by comma\nformat: <postgresql_type>:<go_import>.<go_type>\nexamples: uuid:github.com/google/uuid.UUID,point:src/model.Point,bytea:string\n")

	flags.BoolP(withORM, "q", false, "generate basic ORM queries")
	flags.StringP(dbWrap, "z", "DBWrap", "name of structs for wrapping ORM queries (works only with flag -q, --gen-orm)")
	flags.Bool(withSearch, false, "generate basic Search queries")
	flags.Bool(withValidation, false, "generate model Validation methods")
	flags.Bool(relaxed, false, "use interface{} type in search filters\n")
	flags.BoolP(keepPK, "k", false, "keep primary key name as is (by default it should be converted to 'ID')")
	flags.StringP(softDelete, "s", "", "field for soft_delete tag\n")

	flags.BoolP(noAlias, "w", false, `do not set 'alias' tag to "t"`)
	flags.BoolP(noDiscard, "d", false, "do not use 'discard_unknown_columns' tag\n")

	flags.StringToStringP(json, "j", map[string]string{"*": "map[string]interface{}"}, "type for json columns\nuse format: table.column=type, separate by comma\nuse asterisk as wildcard in table name")
	flags.Bool(jsonTag, false, "add json tag to annotations")

	return
}

// ReadFlags reads basic flags from command
func ReadFlags(command *cobra.Command, o *Options) (err error) {
	var customTypesStrings []string
	uuid := false

	flags := command.Flags()

	if o.URL, err = flags.GetString(Conn); err != nil {
		return
	}

	if o.Output, err = flags.GetString(Output); err != nil {
		return
	}

	if o.Package, err = flags.GetString(Pkg); err != nil {
		return
	}

	if strings.Trim(o.Package, " ") == "" {
		o.Package = path.Base(path.Dir(o.Output))
	}

	if o.Tables, err = flags.GetStringSlice(Tables); err != nil {
		return
	}

	if o.FollowFKs, err = flags.GetBool(FollowFKs); err != nil {
		return
	}

	if o.WithORM, err = flags.GetBool(withORM); err != nil {
		return
	}
	if o.DBWrapName, err = flags.GetString(dbWrap); err != nil {
		return
	}

	if o.WithSearch, err = flags.GetBool(withSearch); err != nil {
		return
	}

	if o.WithValidation, err = flags.GetBool(withValidation); err != nil {
		return
	}
	if o.Relaxed, err = flags.GetBool(relaxed); err != nil {
		return
	}

	if customTypesStrings, err = flags.GetStringSlice(customTypesFlag); err != nil {
		return
	}

	if o.CustomTypes, err = model.ParseCustomTypes(customTypesStrings); err != nil {
		return
	}
	if uuid, err = flags.GetBool(uuidFlag); err != nil {
		return
	}

	if uuid && !o.CustomTypes.Has(model.TypePGUuid) {
		o.CustomTypes.Add(model.TypePGUuid, "uuid.UUID", "github.com/google/uuid")
	}

	if o.KeepPK, err = flags.GetBool(keepPK); err != nil {
		return err
	}

	if o.SoftDelete, err = flags.GetString(softDelete); err != nil {
		return err
	}

	if o.NoDiscard, err = flags.GetBool(noDiscard); err != nil {
		return err
	}

	if o.NoAlias, err = flags.GetBool(noAlias); err != nil {
		return err
	}

	if o.JSONTypes, err = flags.GetStringToString(json); err != nil {
		return err
	}

	if o.AddJSONTag, err = flags.GetBool(jsonTag); err != nil {
		return err
	}

	return
}

// Generate runs whole generation process
func (g Generator) Generate(tables []string, followFKs, useSQLNulls bool, output, tmpl string, packer Packer, customTypes model.CustomTypeMapping) error {
	entities, err := g.Read(tables, followFKs, useSQLNulls, customTypes)
	if err != nil {
		return fmt.Errorf("read database error: %w", err)
	}
	return g.GenerateFromEntities(entities, output, tmpl, packer)
}

func (g Generator) GenerateFromEntities(entities []model.Entity, output, tmpl string, packer Packer) error {
	parsed, err := template.New("base").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing template error: %w", err)
	}

	pack, err := packer(entities)
	if err != nil {
		return fmt.Errorf("packing data error: %w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.ExecuteTemplate(&buffer, "base", pack); err != nil {
		return fmt.Errorf("processing model template error: %w", err)
	}

	saved, err := util.FmtAndSave(buffer.Bytes(), output)

	if err != nil {
		if !saved {
			return fmt.Errorf("saving file error: %w", err)
		}
		log.Printf("formatting file %s error: %s", output, err)
	}
	names := []string{}

	for i, n := range entities {
		names = append(names, n.GoName)
		if i >= 3 {
			break
		}
	}
	fmt.Printf("[%s] Generated %d entitie(s): %30s\n", g.Name, len(entities), output)

	return nil
}

// CreateCommand creates cobra command
func CreateCommand(name, description string, generator Gen) *cobra.Command {
	command := &cobra.Command{
		Use:   name,
		Short: description,
		Long:  "",
		Run: func(command *cobra.Command, args []string) {
			if !command.HasFlags() {
				if err := command.Help(); err != nil {
					log.Printf("help not found, error: %s", err)
				}
				os.Exit(0)
				return
			}

			if err := generator.ReadFlags(command); err != nil {
				log.Printf("read flags error: %s", err)
				return
			}

			if err := generator.Generate(); err != nil {
				log.Printf("generate error: %s", err)
				return
			}
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}

	generator.AddFlags(command)

	return command
}
