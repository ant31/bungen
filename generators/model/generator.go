package model

import (
	"path/filepath"
	"strings"

	"github.com/ant31/bungen/generators/base"
	"github.com/ant31/bungen/model"

	"github.com/spf13/cobra"
)

const (
	keepPK     = "keep-pk"
	noDiscard  = "no-discard"
	noAlias    = "no-alias"
	softDelete = "soft-delete"
	json       = "json"
	jsonTag    = "json-tag"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("model", "Basic bun[postgres] model generator", New())
}

// Basic represents basic generator
type Basic struct {
	options Options
}

// New creates basic generator
func New() *Basic {
	return &Basic{}
}

// Options gets options
func (g *Basic) Options() Options {
	return g.options
}

// SetOptions sets options
func (g *Basic) SetOptions(options Options) {
	g.options = options
}

// AddFlags adds flags to command
func (g *Basic) AddFlags(command *cobra.Command) {
	base.AddFlags(command)

	flags := command.Flags()
	flags.SortFlags = false

	flags.BoolP(keepPK, "k", false, "keep primary key name as is (by default it should be converted to 'ID')")
	flags.StringP(softDelete, "s", "", "field for soft_delete tag\n")

	flags.BoolP(noAlias, "w", false, `do not set 'alias' tag to "t"`)
	flags.BoolP(noDiscard, "d", false, "do not use 'discard_unknown_columns' tag\n")

	flags.StringToStringP(json, "j", map[string]string{"*": "map[string]interface{}"}, "type for json columns\nuse format: table.column=type, separate by comma\nuse asterisk as wildcard in table name")
	flags.Bool(jsonTag, false, "add json tag to annotations")
}

// ReadFlags read flags from command
func (g *Basic) ReadFlags(command *cobra.Command) error {
	var err error

	g.options.URL, g.options.Output, g.options.Package, g.options.Tables, g.options.FollowFKs, g.options.GenORM, g.options.DBWrapName, g.options.CustomTypes, err = base.ReadFlags(command)
	if err != nil {
		return err
	}

	flags := command.Flags()

	if g.options.KeepPK, err = flags.GetBool(keepPK); err != nil {
		return err
	}

	if g.options.SoftDelete, err = flags.GetString(softDelete); err != nil {
		return err
	}

	if g.options.NoDiscard, err = flags.GetBool(noDiscard); err != nil {
		return err
	}

	if g.options.NoAlias, err = flags.GetBool(noAlias); err != nil {
		return err
	}

	if g.options.JSONTypes, err = flags.GetStringToString(json); err != nil {
		return err
	}

	if g.options.AddJSONTag, err = flags.GetBool(jsonTag); err != nil {
		return err
	}

	// setting defaults
	g.options.Def()

	return nil
}

// Generate runs whole generation process
func (g *Basic) Generate() error {
	gen := base.NewGenerator(g.options.URL, "Tables")
	err := gen.Generate(
		g.options.Tables,
		g.options.FollowFKs,
		g.options.UseSQLNulls,
		filepath.Join(g.options.Output, "tables.gen.go"),
		TemplateTable,
		g.Packer(),
		g.options.CustomTypes,
	)
	if err != nil {
		return err
	}
	gen = base.NewGenerator(g.options.URL, "Models")
	entities, err := gen.Read(g.options.Tables,
		g.options.FollowFKs,
		g.options.UseSQLNulls,
		g.options.CustomTypes)
	if err != nil {
		return err
	}

	for i, ent := range entities {
		err = gen.GenerateFromEntities(entities[i:i+1],
			filepath.Join(g.options.Output, strings.ToLower(ent.GoName))+".m.gen.go",
			TemplateModel,
			g.Packer())
		if err != nil {
			return err
		}
	}
	return nil

}

// Packer returns packer function for compile entities into package
func (g *Basic) Packer() base.Packer {
	return func(entities []model.Entity) (interface{}, error) {
		return NewTemplatePackage(entities, g.options), nil
	}
}
