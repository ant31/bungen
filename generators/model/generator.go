package model

import (
	"path/filepath"
	"strings"

	"github.com/ant31/bungen/generators/base"
	"github.com/ant31/bungen/generators/model/templates"
	"github.com/ant31/bungen/model"

	"github.com/spf13/cobra"
)

// CreateCommand creates generator command
func CreateCommand() *cobra.Command {
	return base.CreateCommand("model", "Basic bun[postgres] model generator", New())
}

type Options = base.Options

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

}

// ReadFlags read flags from command
func (g *Basic) ReadFlags(command *cobra.Command) error {
	err := base.ReadFlags(command, &g.options)
	if err != nil {
		return err
	}
	// setting defaults
	g.options.Def()
	return nil
}

func (g *Basic) genPerEntities(name string, tpl string, fileExt string) error {
	gen := base.NewGenerator(g.options.URL, name)
	entities, err := gen.Read(g.options.Tables,
		g.options.FollowFKs,
		g.options.UseSQLNulls,
		g.options.CustomTypes)
	if err != nil {
		return err
	}

	for i, ent := range entities {
		err = gen.GenerateFromEntities(entities[i:i+1],
			filepath.Join(g.options.Output,
				strings.ToLower(ent.GoName))+fileExt+".gen.go",
			tpl,
			g.Packer())
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Basic) genOnce(name string, tpl string, filename string) error {
	gen := base.NewGenerator(g.options.URL, "Tables")
	return gen.Generate(
		g.options.Tables,
		g.options.FollowFKs,
		g.options.UseSQLNulls,
		filepath.Join(g.options.Output, filename),
		tpl,
		g.Packer(),
		g.options.CustomTypes,
	)
}

// Generate runs whole generation process
func (g *Basic) Generate() error {

	err := g.genOnce("Tables", templates.Tables, "tables.gen.go")
	if err != nil {
		return err
	}
	e := ""
	if g.options.WithSearch {
		e += " +search"
		err := g.genOnce("Search", templates.Search, "search.gen.go")
		if err != nil {
			return err
		}
	}

	if g.options.WithORM {
		e += " +orm"
		err := g.genOnce("ORM", templates.ORM, "orm.gen.go")
		if err != nil {
			return err
		}
	}

	err = g.genPerEntities("Models"+e, templates.Model, ".model")
	if err != nil {
		return err
	}

	return nil

}

// Packer returns packer function for compile entities into package
func (g *Basic) Packer() base.Packer {
	return func(entities []model.Entity) (interface{}, error) {
		return NewTemplatePackage(entities, g.options), nil
	}
}
