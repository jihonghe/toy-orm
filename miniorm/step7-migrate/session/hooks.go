package session

import (
	"reflect"

	"miniorm/ormlog"
)

const (
	BeforeQuery  = "BeforeQuery"
	AfterQuery   = "AfterQuery"
	BeforeInsert = "BeforeInsert"
	AfterInsert  = "AfterInsert"
	BeforeUpdate = "BeforeUpdate"
	AfterUpdate  = "AfterUpdate"
	BeforeDelete = "BeforeDelete"
	AfterDelete  = "AfterDelete"
)

func (s *Session) CallHook(method string, tableIns interface{}) {
	table, err := s.RefTable()
	if err != nil {
		ormlog.Error(err)
		return
	}
	hookFn := reflect.ValueOf(table.Model).MethodByName(method)
	if tableIns != nil {
		hookFn = reflect.ValueOf(tableIns).MethodByName(method)
	}
	if !hookFn.IsValid() {
		return
	}

	param := []reflect.Value{reflect.ValueOf(s)}
	returns := hookFn.Call(param)
	if len(returns) > 0 {
		if err, ok := returns[0].Interface().(error); ok {
			ormlog.Error(err)
			return
		}
	}
	return
}
