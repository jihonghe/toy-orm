package session

import (
	"reflect"

	"miniorm/clause"
	"miniorm/schema"
)

// Insert will insert records given by the instance of table struct
func (s *Session) Insert(values ...interface{}) (rowsAffected int64, err error) {
	var recordValues []interface{}
	var refTable *schema.Schema
	for _, value := range values {
		refTable, err = s.Model(value).RefTable()
		if err != nil {
			return
		}
		s.clause.Set(clause.INSERT, refTable.Name, refTable.FieldNames)
		recordValues = append(recordValues, refTable.Struct2Value(value))
	}
	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return
	}

	return result.RowsAffected()
}

// Find will set the records queried from database to the instance of table struct
func (s *Session) Find(values interface{}) (err error) {
	dstSlc := reflect.Indirect(reflect.ValueOf(values))
	dstType := dstSlc.Type().Elem()
	refTable, err := s.Model(reflect.New(dstType).Elem().Interface()).RefTable()
	if err != nil {
		return
	}

	s.clause.Set(clause.SELECT, refTable.Name, refTable.FieldNames)
	sqlClause, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sqlClause, vars...).QueryRows()
	if err != nil {
		return
	}

	for rows.Next() {
		dst := reflect.New(dstType).Elem()
		var fields []interface{}
		for _, name := range refTable.FieldNames {
			fields = append(fields, dst.FieldByName(name).Addr().Interface())
		}
		if err = rows.Scan(fields...); err != nil {
			return
		}
		dstSlc.Set(reflect.Append(dstSlc, dst))
	}

	return rows.Close()
}
