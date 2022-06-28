package session

import (
	"database/sql"
	"strings"

	"miniorm/dialect"
	"miniorm/ormlog"
	"miniorm/schema"
)

type Session struct {
	db       *sql.DB         // database conn instance
	dialect  dialect.Dialect // the database type of this session connected
	refTable *schema.Schema  // the table of this session operates
	sql      strings.Builder // use strings.Builder to avoid memory allocation when build sql
	sqlVars  []interface{}   // the vars in sql placeholder
}

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{db: db, dialect: dialect}
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
}

func (s *Session) DB() *sql.DB {
	return s.db
}

func (s *Session) Exec() (res sql.Result, err error) {
	defer s.Clear()
	ormlog.Debug(s.sql.String(), s.sqlVars)
	if res, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		ormlog.Error(err)
	}

	return
}

// QueryRow get a record from table in session
func (s *Session) QueryRow() (row *sql.Row) {
	defer s.Clear()
	ormlog.Debug(s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

// QueryRows get the rows of a query
//  NOTES: sql.Rows is usually used for method QueryRows, and QueryRow returns sql.Row
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	ormlog.Debug(s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		ormlog.Error(err)
	}

	return
}

// Raw get a session by raw sql
func (s *Session) Raw(sql string, values ...interface{}) (session *Session) {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}
