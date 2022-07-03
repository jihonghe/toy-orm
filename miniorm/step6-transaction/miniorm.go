package miniorm

import (
	"database/sql"
	"fmt"

	"miniorm/dialect"
	"miniorm/ormlog"
	"miniorm/session"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewEngine(driver, dataSource string) (e *Engine, err error) {
	db, err := sql.Open(driver, dataSource)
	if err != nil {
		ormlog.Error(err)
		return
	}
	if err = db.Ping(); err != nil {
		ormlog.Error(err)
		return
	}
	// make sure the specific dialect exists
	dial, ok := dialect.GetDialect(driver)
	if !ok {
		return nil, ormlog.New(fmt.Sprintf("dialect %s NOT FOUND", driver))
	}
	e = &Engine{db: db, dialect: dial}
	ormlog.Info("database connected")
	return
}

func (e *Engine) Close() (err error) {
	err = e.db.Close()
	return
}

func (e *Engine) NewSession() (s *session.Session) {
	return session.New(e.db, e.dialect)
}
