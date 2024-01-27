package modelgen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/thorn-jmh/errorst"
)

type Decl interface {
	Gen(file *jen.File) error
}

func (d *Object) Gen(f *jen.File) error {
	// first declare struct fields
	var fieldsDecl = func(g *jen.Group) {
		// add gorm model
		//declGormModel(g)

		// decl fields
		for _, field := range d.Fields {
			stat := declType(g.Id(field.Name), field.Type)
			if field.Tags != nil {
				stat.Tag(field.Tags)
			}
			if field.Comment != "" {
				stat.Comment(field.Comment)
			}
		}

	}

	// second declare sub relation references
	var subRelationsDecl = func(g *jen.Group) {
		fieldsDecl(g)
		if len(d.SubRelations) > 0 {
			g.Line().Commentf("sub relations of %s", d.Name)
			for _, sub := range d.SubRelations {
				g.Id(sub.Name + "s").Index().Id(sub.Name)
			}
		}
	}

	// third declare struct
	f.Line().Comment(d.Comment)
	f.Type().Id(d.Name).StructFunc(subRelationsDecl)

	// forth declare definitions
	for _, def := range d.Definitions {
		if err := def.Gen(f); err != nil {
			return errorst.Wrap(err, "failed to generate definition<%s>", def)
		}
	}

	// fifth declare sub relations
	for _, sub := range d.SubRelations {
		if err := sub.Gen(f); err != nil {
			return errorst.Wrap(err, "failed to generate sub relation<%s>", sub)
		}
	}

	return nil
}

func (d *Alias) Gen(f *jen.File) error {
	// just alias it
	f.Line().Comment(d.Comment)
	declType(f.Type().Id(d.Name), d.BaseType)
	return nil
}

func (d *Enum) Gen(f *jen.File) error {
	// first alias it
	if err := d.Alias.Gen(f); err != nil {
		return errorst.Wrap(err, "failed to alias enum<%s>", d.Name)
	}

	// second declare values
	f.Line().Commentf("enum %s values", d.Name)
	var enumValues []jen.Code
	for _, value := range d.Values {
		ev := jen.Id(fmt.Sprintf("%s_%v", d.Name, value)).Id(d.Name).Op("=").Lit(value)
		enumValues = append(enumValues, ev)
	}
	f.Const().Defs(enumValues...)

	return nil
}

func declType(s *jen.Statement, typ Type) *jen.Statement {
	if typ.NilAble {
		s.Op("*")
	}
	if typ.Domain != "" {
		return s.Qual(typ.Domain, typ.Name)
	} else {
		return s.Id(typ.Name)
	}
}

func declGormModel(g *jen.Group) *jen.Statement {
	return g.Id("").Qual("gorm.io/gorm", "Model")
}
