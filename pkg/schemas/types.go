package schemas

type SchemaNodeType string

const (
	TypeNameString  SchemaNodeType = "string"
	TypeNameArray   SchemaNodeType = "array"
	TypeNameNumber  SchemaNodeType = "number"
	TypeNameInteger SchemaNodeType = "integer"
	TypeNameObject  SchemaNodeType = "object"
	TypeNameBoolean SchemaNodeType = "boolean"
	TypeNameNull    SchemaNodeType = "null"
)

// IsPrimitiveType returns true if the given nodeType is a primitive type.
// https://json-schema.org/draft/2020-12/json-schema-core#section-4.2.1
func IsPrimitiveType(t SchemaNodeType) bool {
	switch t {
	case TypeNameString, TypeNameNumber, TypeNameInteger, TypeNameBoolean, TypeNameNull:
		return true
	default:
		return false
	}
}
