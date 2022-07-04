package session

import (
	"database/sql"
	"strings"

	"miniorm/clause"
	"miniorm/dialect"
	"miniorm/ormlog"
	"miniorm/schema"
)

var (
	_ CommonDB = (*sql.DB)(nil)
	_ CommonDB = (*sql.Tx)(nil)
)

// CommonDB is the minimal intersected function set of sql.DB and sql.Tx
//  Use CommonDB as the abstraction sql.DB and sql.Tx
//  Why Query, QueryRow and Exec? Because they are called by others DB operation func like Update, Insert and so on
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Session struct {
	db       *sql.DB         // database conn instance
	tx       *sql.Tx         // for transaction, it means open transaction when it is not nil
	dialect  dialect.Dialect // the database type of this session connected
	refTable *schema.Schema  // the table of this session operates
	clause   clause.Clause   // build the complete sql statement
	sql      strings.Builder // use strings.Builder to avoid memory allocation when build sql
	sqlVars  []interface{}   // the vars in sql placeholder
}

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{db: db, dialect: dialect}
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}

// DB returns *sql.Tx if a tx begins, otherwise returns *sql.DB
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
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
