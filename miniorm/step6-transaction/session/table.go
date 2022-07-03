package session

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"miniorm/ormlog"
	"miniorm/schema"
)

// Model parses the given param 'v' to the dialect of Session
func (s *Session) Model(v interface{}) (session *Session) {
	if s.refTable == nil || reflect.TypeOf(v) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Parse(v, s.dialect)
	}
	return s
}

// RefTableName returns the table name in session, it returns "" if Session.refTable is nil
func (s *Session) RefTableName() (tableName string) {
	if s.refTable != nil {
		return s.refTable.Name
	}
	return
}

func (s *Session) RefTable() (schema *schema.Schema, err error) {
	if s.refTable == nil {
		return nil, ormlog.New("Model in session is not set")
	}
	return s.refTable, nil
}

func (s *Session) CreateTable() (err error) {
	table, err := s.RefTable()
	if err != nil {
		return err
	}
	var columns []string
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Constraints))
	}
	columnsDesc := strings.Join(columns, ",")
	_, err = s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.Name, columnsDesc)).Exec()

	return
}

func (s *Session) DropTable() (err error) {
	_, err = s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s;", s.RefTableName())).Exec()
	return
}

func (s *Session) TableExists() (exist bool, err error) {
	existSQL, sqlVars := s.dialect.TableExistSQL(s.RefTableName())
	row := s.Raw(existSQL, sqlVars...).QueryRow()
	var tableName string
	err = row.Scan(&tableName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return
	}
	return tableName == s.RefTableName(), nil
}
