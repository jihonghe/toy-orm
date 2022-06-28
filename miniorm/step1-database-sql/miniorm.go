package miniorm

import (
	"database/sql"

	"miniorm/ormlog"
	"miniorm/session"
)

type Engine struct {
	db *sql.DB
}

func NewEngine(driver, dataSourceName string) (e *Engine, err error) {
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		ormlog.Error(err)
		return
	}
	if err = db.Ping(); err != nil {
		ormlog.Error(err)
		return
	}
	e = &Engine{db: db}
	ormlog.Info("database connected")
	return
}

func (e *Engine) Close() (err error) {
	err = e.db.Close()
	return
}

func (e *Engine) NewSession() (s *session.Session) {
	return session.New(e.db)
}
