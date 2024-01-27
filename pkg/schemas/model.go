package schemas

import (
	"encoding/json"
	"github.com/thorn-jmh/errorst"
)

// Definitions hold schema definitions.
type Definitions map[string]*SubSchema

// Schema is the root schema.
type Schema struct {
	// Root schema infos
	// https://json-schema.org/draft/2020-12/json-schema-core
	Definitions Definitions `json:"$defs,omitempty"`   // #section-8.2.4
	Version     string      `json:"$schema,omitempty"` // #section-8.1
	*SubSchema
}

// Type is a list of type names.
// https://json-schema.org/draft/2020-12/json-schema-validation#name-type
// After parsing, we use a slice of strings to represent node type.
type Type []SchemaNodeType

func (t *Type) Contains(typ SchemaNodeType) bool {
	for _, t := range *t {
		if t == typ {
			return true
		}
	}
	return false
}

// Value is a JSON value.
type Value any

type SubSchema SchemaProperties

// SchemaProperties represents a JSON Schema object type.
// A subSchema could be a boolean value.
//
// NOTE: for all fields whose default value behavior is NOT
// same with the null value, we use pointer type.
type SchemaProperties struct {
	// ID and Reference
	// https://json-schema.org/draft/2020-12/json-schema-core
	ID  string `json:"$id"`            // #section-8.2.1
	Ref string `json:"$ref,omitempty"` // #section-8.2.3

	// Meta-Data
	// https://json-schema.org/draft/2020-12/json-schema-validation#section-9
	Title       string  `json:"title,omitempty"`       // #section-9.1
	Description string  `json:"description,omitempty"` // #section-9.1
	Default     Value   `json:"default,omitempty"`     // #section-9.2
	Deprecated  bool    `json:"deprecated,omitempty"`  // #section-9.3
	ReadOnly    bool    `json:"readOnly,omitempty"`    // #section-9.4
	WriteOnly   bool    `json:"writeOnly,omitempty"`   // #section-9.4 TODO: 这玩意有啥用x
	Examples    []Value `json:"examples,omitempty"`    // #section-9.5

	// Validation
	// https://json-schema.org/draft/2020-12/json-schema-validation
	// For any instance.
	Type  Type    `json:"type,omitempty"`  // #section-6.1.1
	Enum  []Value `json:"enum,omitempty"`  // #section-6.1.2
	Const Value   `json:"const,omitempty"` // #section-6.1.3
	// For numeric instance.
	MultipleOf       *float64 `json:"multipleOf,omitempty"`       // #section-6.2.1
	Maximum          *float64 `json:"maximum,omitempty"`          // #section-6.2.2
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"` // #section-6.2.3
	Minimum          *float64 `json:"minimum,omitempty"`          // #section-6.2.4
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"` // #section-6.2.5
	// For string
	MaxLength *int   `json:"maxLength,omitempty"` // #section-6.3.1
	MinLength *int   `json:"minLength,omitempty"` // #section-6.3.2
	Pattern   string `json:"pattern,omitempty"`   // #section-6.3.3
	// For array
	MaxItems    *int `json:"maxItems,omitempty"`    // #section-6.4.1
	MinItems    *int `json:"minItems,omitempty"`    // #section-6.4.2
	UniqueItems bool `json:"uniqueItems,omitempty"` // #section-6.4.3
	MaxContains *int `json:"maxContains,omitempty"` // #section-6.4.4
	MinContains *int `json:"minContains,omitempty"` // #section-6.4.5
	// For object
	MaxProperties     *int                `json:"maxProperties,omitempty"`     // #section-6.5.1
	MinProperties     *int                `json:"minProperties,omitempty"`     // #section-6.5.2
	Required          []string            `json:"required,omitempty"`          // #section-6.5.3
	DependentRequired map[string][]string `json:"dependentRequired,omitempty"` // #section-6.5.4
	// Format
	Format string `json:"format,omitempty"` // #sectino-7

	// Applying subSchema
	// https://json-schema.org/draft/2020-12/json-schema-core#section-10
	// Logic
	AllOf []*SubSchema `json:"allOf,omitempty"` // #section-10.2.1.1
	AnyOf []*SubSchema `json:"anyOf,omitempty"` // #section-10.2.1.2
	OneOf []*SubSchema `json:"oneOf,omitempty"` // #section-10.2.1.3
	Not   *SubSchema   `json:"not,omitempty"`   // #section-10.2.1.4
	// For array
	PrefixItems []*SubSchema `json:"prefixItems,omitempty"` // #section-10.3.1.1
	Items       *SubSchema   `json:"items,omitempty"`       // #section-10.3.1.2
	Contains    *SubSchema   `json:"contains,omitempty"`    // #section-10.3.1.3
	// For object
	Properties           map[string]*SubSchema `json:"properties,omitempty"`           // #section-10.3.2.1
	PatternProperties    map[string]*SubSchema `json:"patternProperties,omitempty"`    // #section-10.3.2.2
	AdditionalProperties *SubSchema            `json:"additionalProperties,omitempty"` // #section-10.3.2.3
	PropertyNames        *SubSchema            `json:"propertyNames,omitempty"`        // #section-10.3.2.4
}

// >>>>>>>>>>>>>>>>>>>> impl UnmarshalJSON >>>>>>>>>>>>>>>>>>>>>>>

// type alias for unmarshal
type schemaToUnmarshal Schema
type subSchemaToUnmarshal SchemaProperties

// UnmarshalJSON implements json.Unmarshaler for Schema struct.
func (s *Schema) UnmarshalJSON(data []byte) error {
	var unmarshalSchema schemaToUnmarshal

	if err := json.Unmarshal(data, &unmarshalSchema); err != nil {
		return errorst.NewError("failed to unmarshal schema: %w", err)
	}

	// Take care of legacy fields.
	// https://json-schema.org/draft-04/draft-zyp-json-schema-04
	var legacySchema struct {
		Definitions Definitions `json:"definitions,omitempty"`
	}
	if err := json.Unmarshal(data, &legacySchema); err != nil {
		return errorst.NewError("failed to unmarshal schema: %w", err)
	}

	// Fall back to definitions if $defs is not present.
	if unmarshalSchema.Definitions == nil {
		unmarshalSchema.Definitions = legacySchema.Definitions
	}

	*s = (Schema)(unmarshalSchema)

	return nil
}

// UnmarshalJSON implements json.Unmarshaler for Type.
func (t *Type) UnmarshalJSON(b []byte) error {
	// if the type is a list, unmarshal it as a list of strings.
	if len(b) > 0 && b[0] == '[' {
		var s []SchemaNodeType
		if err := json.Unmarshal(b, &s); err != nil {
			return errorst.NewError("failed to unmarshal type list: %w", err)
		}
		*t = s
		return nil
	}

	// else unmarshal it as a single string.
	var s SchemaNodeType
	if err := json.Unmarshal(b, &s); err != nil {
		return errorst.NewError("failed to unmarshal type: %w", err)
	}
	if s != "" {
		*t = []SchemaNodeType{s}
	} else {
		*t = nil
	}

	return nil
}

// UnmarshalJSON accepts booleans as schemas where `true` is equivalent to `{}`
// and `false` is equivalent to `{"not": {}}`.
func (value *SchemaProperties) UnmarshalJSON(raw []byte) error {
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		if b {
			*value = SchemaProperties{}
		} else {
			*value = SchemaProperties{Not: &SubSchema{}}
		}

		return nil
	}

	var obj subSchemaToUnmarshal
	if err := json.Unmarshal(raw, &obj); err != nil {
		return errorst.NewError("failed to unmarshal subSchema: %w", err)
	}

	// Take care of legacy fields from older RFC versions.
	// https://json-schema.org/draft-04/draft-zyp-json-schema-04
	legacySubSchema := struct {
		ID string `json:"id"`
	}{}
	if err := json.Unmarshal(raw, &legacySubSchema); err != nil {
		return errorst.NewError("failed to unmarshal subSchema: %w", err)
	}
	if obj.ID == "" {
		obj.ID = legacySubSchema.ID
	}

	*value = SchemaProperties(obj)

	return nil
}
