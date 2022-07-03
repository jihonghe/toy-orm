package session

import (
	"database/sql"
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
	sqlClause, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sqlClause, vars...).Exec()
	if err != nil {
		return
	}

	return result.RowsAffected()
}

func (s *Session) First(value interface{}) (err error) {
	dst := reflect.Indirect(reflect.ValueOf(value))
	dstSlc := reflect.New(reflect.SliceOf(dst.Type())).Elem()
	if err = s.Limit(0, 1).Find(dstSlc.Addr().Interface()); err != nil {
		return
	}
	if dstSlc.Len() == 0 {
		return sql.ErrNoRows
	}
	dst.Set(dstSlc.Index(0))
	return
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
	// NOTES: in the SELECT clause, add WHERE, ORDERBY and LIMIT in order whether it exists or not
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

// Update
//  param1: supports 2 format type:
//      1.map[string][]interface{}, key: condition-desc, value: values for condition-desc
//      2.key-value pairs, it will be converted to map[string]interface{}, example: Update("Name", "Tom", "Age", 11)
func (s *Session) Update(kv ...interface{}) (rowsAffected int64, err error) {
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	s.clause.Set(clause.UPDATE, s.RefTableName(), m)
	// NOTES: In order to build the correct sequence, add clause.WHERE in the end whether it exists or not
	sqlClause, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sqlClause, vars...).Exec()
	if err != nil {
		return
	}

	return result.RowsAffected()
}

func (s *Session) Delete() (rowsAffected int64, err error) {
	s.clause.Set(clause.DELETE, s.RefTableName())
	// NOTES: In order to build the correct sequence, add clause.WHERE in the end whether it exists or not
	sqlClause, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sqlClause, vars...).Exec()
	if err != nil {
		return
	}
	return result.RowsAffected()
}

func (s *Session) Count() (count int64, err error) {
	s.clause.Set(clause.COUNT, s.RefTableName())
	// NOTES: In order to build the correct sequence, add clause.WHERE in the end whether it exists or not
	sqlClause, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sqlClause, vars...).QueryRow()
	if err != nil {
		return
	}
	if err = row.Scan(&count); err != nil {
		return
	}

	return
}

// Where
//  if there are multiple WHERE clause in the chain, only the last one takes effect
func (s *Session) Where(desc string, args ...interface{}) (session *Session) {
	s.clause.Set(clause.WHERE, append(append([]interface{}{}, desc), args...)...)
	return s
}

// Limit
//  if there are multiple LIMIT clause in the chain, only the last one takes effect
func (s *Session) Limit(offset, limit uint64) (session *Session) {
	s.clause.Set(clause.LIMIT, offset, limit)
	return s
}

// OrderBy
//  if there are multiple ORDERBY clause in the chain, only the last one takes effect
func (s *Session) OrderBy(desc string) (session *Session) {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}
