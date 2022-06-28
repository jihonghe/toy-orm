package session

import (
	"database/sql"
	"strings"

	"miniorm/ormlog"
)

type Session struct {
	db      *sql.DB         // database conn instance
	sql     strings.Builder // use strings.Builder to avoid memory allocation when build sql
	sqlVars []interface{}
}

func New(db *sql.DB) *Session {
	return &Session{db: db}
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
