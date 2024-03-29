package modelgen

import (
	"dbgen/pkg/schemas"
	"github.com/thorn-jmh/errorst"
	"net/url"
	"strings"
)

var MainSchema *schemas.Schema

func GenAndProcess(sch *schemas.Schema) (*Object, error) {
	// first: generate object
	obj, err := GenerateModel(sch)
	if err != nil {
		return nil, errorst.Wrap(err, "failed to generate model")
	}

	// second: process object
	err = ProcessTree(obj)
	if err != nil {
		return nil, errorst.Wrap(err, "failed to process model")
	}

	// third: process association
	ProcessAssociation(obj, obj.Name)

	return obj, nil
}

func ProcessAssociation(obj *Object, parentName string) {
	// if Name equal to parent name, this is a database schema
	// add ID field
	if obj.Name == parentName {
		// add ID field to root obj
		idField := Field{
			Name: "ID",
			Type: Type{
				Name: "uint",
			},
			Tags: map[string]string{
				"json": "-",
				"gorm": "primaryKey",
			},
		}
		obj.Fields = append(obj.Fields, idField)
	}

	for _, def := range obj.Definitions {
		if defObj, ok := def.(*Object); ok {
			ProcessAssociation(defObj, parentName)
		}
	}

	// add association foreignKey to obj
	for _, sub := range obj.SubRelations {
		// add association foreignKey to subRelation
		foreignKeyField := Field{
			Name: parentName + "ID",
			Type: Type{
				Name: "uint",
			},
			Tags: map[string]string{
				"json": "-",
			},
			Comment: "foreign key to " + obj.Name,
		}
		sub.Fields = append(sub.Fields, foreignKeyField)

		ProcessAssociation(sub, sub.Name)
	}
}

func ProcessTree(obj *Object) (err error) {

	// first: process sub relations
	for _, sub := range obj.SubRelations {
		if !isNamedObject(sub) {
			// we do not accept unnamed
			return errorst.Wrap(ErrInvalidStructure, "unnamed sub relation: %+v", sub)
		}
		if err := ProcessTree(sub); err != nil {
			return errorst.Wrap(err, "failed to process sub relation<%s>", sub.Name)
		}
	}

	// second: process definitions
	var newDefinitions []Decl
	for _, def := range obj.Definitions {
		// deep search
		if defObj, ok := def.(*Object); ok {
			if err := ProcessTree(defObj); err != nil {
				return errorst.Wrap(err, "failed to process object<%s>", defObj.Name)
			}

			// if child is not named, append all fields and definitions and sub relations, then delete it
			if !isNamedObject(defObj) {
				obj.Fields = append(obj.Fields, defObj.Fields...)
				obj.SubRelations = append(obj.SubRelations, defObj.SubRelations...)

				newDefinitions = append(newDefinitions, defObj.Definitions...)
			} else {
				newDefinitions = append(newDefinitions, def)
			}
		} else {
			newDefinitions = append(newDefinitions, def)
		}
	}
	obj.Definitions = newDefinitions

	return nil
}

func GenerateModel(sch *schemas.Schema) (obj *Object, err error) {
	// first check if the schema is an object type
	if !isObjectType(sch.Type) {
		return nil, errorst.Wrap(ErrWrongSyntax, "Invalid main schema type: %+v", sch.Type)
	}

	// second: main schema
	MainSchema = sch

	// third: generate object
	return GenerateObject(Context{
		State{
			Path: sch.ID + "#",
		},
	}, sch.SubSchema)
}

func GenerateObject(ctx Context, sch *schemas.SubSchema) (obj *Object, err error) {
	obj = &Object{}

	// check if the schema is an object type
	if !isObjectType(sch.Type) {
		if isPrimitiveType(sch.Type) {
			return GeneratePrimitive(ctx, sch)
		} else if isArrayType(sch.Type) {
			return GenerateArray(ctx, sch)
		} else if sch.Ref != "" {
			return GenerateRef(ctx, sch)
		} else {
			return nil, errorst.Wrap(ErrWrongSyntax, "Invalid schema type: %+v", sch.Type)
		}

	}

	// first: process meta-data
	if name, err := path2Name(ctx.Path); err != nil {
		return nil, errorst.Wrap(err, "failed to get object name at %s", ctx.Path)
	} else {
		obj.Name = name
		obj.Comment = getComment(sch)
	}

	// second: check ref  >>> DELETE for we can not name it
	//if sch.Ref != "" {
	//
	//}

	// third: process properties
	for pName, pSch := range sch.Properties {
		newCtx := Context{
			State{
				Require: isRequired(pName, sch),
				Path:    ctx.Path + "/" + pName,
			},
		}

		// get property object and add 2 definitions
		pObj, err := GenerateObject(newCtx, pSch)
		if err != nil {
			return nil, errorst.Wrap(err, "failed to generate object <%s> at %s", pName, ctx.Path)
		}
		obj.Definitions = append(obj.Definitions, pObj)

		// if it's a named object, add it to field
		if isNamedObject(pObj) {
			pTyp := Type{
				Name: pObj.Name,
			}
			pTyp.NilAble = isRequired(pName, sch)

			field := Field{
				Name: BigCamelStyle(pName),
				Type: pTyp,
				Tags: make(map[string]string),
			}
			setFieldJsonTag(&field, pName)
			setFieldGormTag(&field, true)

			obj.Fields = append(obj.Fields, field)
		}

	}
	return
}

func GeneratePrimitive(ctx Context, sch *schemas.SubSchema) (obj *Object, err error) {
	obj = &Object{}

	// first: get primitive type
	typ, err := getPrimitiveType(sch)
	if err != nil {
		return nil, errorst.Wrap(err, "failed to get primitive type at %s", ctx.Path)
	}
	if ctx.Require {
		typ.NilAble = false
	}

	// second: if enums
	if sch.Enum != nil && len(sch.Enum) > 0 {
		// create type alias
		name, err := path2Name(ctx.Path)
		if err != nil {
			return nil, errorst.Wrap(err, "failed to get enum name at %s", ctx.Path)
		}
		alias := Alias{
			Name: name,
			BaseType: Type{
				Name:   typ.Name,
				Domain: typ.Domain, // we do not care whether inner type is nilAble
			},
		}

		// create enum and add it to definitions
		enum := Enum{
			Alias: alias,
		}
		addValue2Enum(&enum, sch.Enum...)
		obj.Definitions = append(obj.Definitions, &enum)

		// institute primitive type
		typ = Type{
			Name:    enum.Name,
			NilAble: typ.NilAble,
		}
	}

	// third: create field of typ
	pathElems := strings.Split(ctx.Path, "/")
	fName := pathElems[len(pathElems)-1]
	field := Field{
		Name:    BigCamelStyle(fName),
		Type:    typ,
		Comment: getComment(sch),
		Tags:    make(map[string]string),
	}
	setFieldJsonTag(&field, fName)
	obj.Fields = append(obj.Fields, field)
	return
}

func GenerateArray(ctx Context, sch *schemas.SubSchema) (obj *Object, err error) {
	obj = &Object{}

	// get array item type
	newCtx := Context{
		State{
			Path: ctx.Path + "/item",
		},
	}
	itemObj, err := GenerateObject(newCtx, sch.Items)
	if err != nil {
		return nil, errorst.Wrap(err, "failed to generate array item at %s", ctx.Path)
	}

	// add 2 sub relations
	obj.SubRelations = append(obj.SubRelations, itemObj)

	// add reference field
	pathElems := strings.Split(ctx.Path, "/")
	fName := pathElems[len(pathElems)-1]
	field := Field{
		Name: BigCamelStyle(fName) + "Items",
		Type: Type{
			Name:    itemObj.Name,
			IsArray: true,
		},
		Comment: getComment(sch),
		Tags:    make(map[string]string),
	}
	setFieldJsonTag(&field, fName)
	obj.Fields = append(obj.Fields, field)
	return
}

func GenerateRef(ctx Context, sch *schemas.SubSchema) (obj *Object, err error) {
	// first: get ref schema
	refSch, err := getRefSchema(sch.Ref)
	if err != nil {
		return nil, errorst.Wrap(err, "failed to get ref schema at %s", ctx.Path)
	}
	return GenerateObject(ctx, refSch)
}

func addValue2Enum(enum *Enum, value ...schemas.Value) {
	for _, v := range value {
		enum.Values = append(enum.Values, v)
	}
}

func getRefSchema(path string) (*schemas.SubSchema, error) {
	// first get def name from path
	uri, err := url.Parse(path)
	if err != nil {
		return nil, errorst.Wrap(err, "failed to parse ref path: %s", path)
	}
	frags := strings.Split(uri.Fragment, "/")
	if len(frags) < 3 || frags[1] != "$defs" {
		return nil, errorst.Wrap(ErrWrongSyntax, "invalid ref path: %s", path)
	}
	defName := frags[2]

	// second get schema from main schema
	if sch, ok := MainSchema.Definitions[defName]; ok {
		return sch, nil
	} else {
		return nil, errorst.Wrap(ErrWrongSyntax, "failed to get ref schema: %s", path)
	}
}

func path2Name(path string) (string, error) {
	uri, err := url.Parse(path)
	if err != nil {
		return "", errorst.Wrap(err, "failed to parse path: %s", path)
	}

	// rule
	// 1. last part of path is schema (remove suffix)
	// 2. fragments is name
	paths := strings.Split(uri.Path, "/")
	if len(paths) == 0 {
		return "", errorst.Wrap(ErrWrongSyntax, "invalid path: %s", path)
	}
	schema := strings.Split(paths[len(paths)-1], ".")[0]
	frags := strings.Split(uri.Fragment, "/")

	// format
	var ret = BigCamelStyle(schema)
	for _, f := range frags {
		ret += BigCamelStyle(f)
	}
	return ret, nil
}

func setFieldJsonTag(field *Field, name string) {
	field.Tags["json"] = name
	if field.Type.NilAble {
		field.Tags["json"] += ",omitempty"
	}
}

func setFieldGormTag(field *Field, isEmbedded bool) {
	if isEmbedded {
		field.Tags["gorm"] = "embedded"
	}
}

func getComment(sch *schemas.SubSchema) string {
	if sch.Title != "" && sch.Description != "" {
		return sch.Title + ": " + sch.Description
	} else if sch.Title != "" {
		return sch.Title
	}
	return sch.Description
}

func isRequired(pName string, sch *schemas.SubSchema) bool {
	for _, r := range sch.Required {
		if r == pName {
			return true
		}
	}
	return false
}

func isNamedObject(obj *Object) bool {
	return obj.Name != ""
}

func isPrimitiveType(typ schemas.Type) bool {
	// contains one of the primitive types
	if typ.Contains(schemas.TypeNameString) || typ.Contains(schemas.TypeNameInteger) || typ.Contains(schemas.TypeNameNumber) || typ.Contains(schemas.TypeNameBoolean) {
		return true
	}
	return false
}

func isArrayType(typ schemas.Type) bool {
	return typ.Contains(schemas.TypeNameArray)
}

func isObjectType(typ schemas.Type) bool {
	return typ.Contains(schemas.TypeNameObject)
}

func isNilAble(typ schemas.Type) bool {
	return typ.Contains(schemas.TypeNameNull)
}

func getPrimitiveType(schema *schemas.SubSchema) (ret Type, err error) {
	if schema.Type.Contains(schemas.TypeNameString) {
		if schema.Format == "date-time" || schema.Format == "date" {
			ret = Type{
				Name:   "Time",
				Domain: "time",
			}
		} else {
			ret = Type{
				Name:   "string",
				Domain: "",
			}
		}
	} else if schema.Type.Contains(schemas.TypeNameInteger) {
		ret = Type{
			Name:   "int",
			Domain: "",
		}
	} else if schema.Type.Contains(schemas.TypeNameNumber) {
		ret = Type{
			Name:   "float64",
			Domain: "",
		}
	} else if schema.Type.Contains(schemas.TypeNameBoolean) {
		ret = Type{
			Name:   "bool",
			Domain: "",
		}
	} else {
		return Type{}, errorst.Wrap(ErrWrongSyntax, "Invalid primitive type: %+v", schema.Type)
	}

	if isNilAble(schema.Type) {
		ret.NilAble = true
	}

	return
}
