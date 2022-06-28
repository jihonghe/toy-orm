package dialect

import (
	"reflect"
)

var (
	dialectsMap = map[string]Dialect{}
)

// Dialect gives 2 methods to get the datatype and check table if exist for different database
type Dialect interface {
	DataTypeOf(typ reflect.Value) (dataType string)
	TableExistSQL(tableName string) (sql string, sqlVars []interface{})
}

func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect
}

func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
