package dialect

import (
	"reflect"
)

// TODO: mysql field type is too much to implement, (Q^Q)!

type mysql struct{}

var _ Dialect = (*mysql)(nil)

func (m *mysql) DataTypeOf(typ reflect.Value) (dataType string) {
	// TODO implement me
	panic("implement me")
}

func (m *mysql) TableExistSQL(tableName string) (sql string, sqlVars []interface{}) {
	// TODO implement me
	panic("implement me")
}
