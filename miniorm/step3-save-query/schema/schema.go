package schema

import (
	"go/ast"
	"reflect"

	"miniorm/dialect"
)

// Field represents a column of database
type Field struct {
	Name        string
	Type        string
	Constraints string // the constraints are parsed from struct field tag 'miniorm'
}

// Schema represents a table of database
type Schema struct {
	Model      interface{}       // the mapping object(pointer instance of Table struct)
	Name       string            // table name
	Fields     []*Field          // columns in table
	FieldNames []string          // column names in table
	fieldMap   map[string]*Field // the mapping of column name and column object, used for get column object by name
}

func (s *Schema) GetField(name string) (field *Field) {
	return s.fieldMap[name]
}

// Parse parses the given model to the specified schema of dialect
func Parse(dst interface{}, dialect dialect.Dialect) (schema *Schema) {
	modelType := reflect.Indirect(reflect.ValueOf(dst)).Type()
	schema = &Schema{
		Model:    dst,
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}
	for i := 0; i < modelType.NumField(); i++ {
		member := modelType.Field(i)
		// skip the struct member which is anonymous and unexported
		if !member.Anonymous && !ast.IsExported(member.Name) {
			continue
		}
		field := &Field{
			Name: member.Name,
			// TODO: figure it out why not use member.Type.String()
			Type: dialect.DataTypeOf(reflect.Indirect(reflect.New(member.Type))),
		}
		if tag, ok := member.Tag.Lookup("miniorm"); ok {
			field.Constraints = tag
		}
		schema.Fields = append(schema.Fields, field)
		schema.FieldNames = append(schema.FieldNames, field.Name)
		schema.fieldMap[field.Name] = field
	}
	return
}
