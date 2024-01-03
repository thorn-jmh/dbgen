package modelgen

// >>>>>>>>>>>> this used to describe the result of model generation >>>>>>>>>>>>>>>

type Object struct {
	// meta data
	Name    string // struct type's name
	Comment string // struct type's comment
	// fields
	Fields []Field // fields of struct type
	// tree structure
	Definitions  []Decl
	SubRelations []*Object
}

type Field struct {
	Name    string            // field name
	Type    Type              // field Type
	Tags    map[string]string // tags of this field
	Comment string            // comment on this field
}

type Alias struct {
	Name     string // alias type's name
	Comment  string // alias type's comment
	BaseType Type   // alias type's base type
}

type Enum struct {
	Alias
	Values []any
}

type Type struct {
	Name    string // type Name
	Domain  string // package path
	NilAble bool   // is NilAble, we will use pointer to represent NilAble type
}
