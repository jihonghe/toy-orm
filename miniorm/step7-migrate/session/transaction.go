package session

import (
	"miniorm/ormlog"
)

func (s *Session) Begin() (err error) {
	ormlog.Info("transaction begin")
	s.tx, err = s.db.Begin()
	return
}

func (s *Session) Commit() (err error) {
	err = s.tx.Commit()
	if err == nil {
		ormlog.Info("transaction commit")
	}
	return
}

func (s *Session) Rollback() (err error) {
	err = s.tx.Rollback()
	if err == nil {
		ormlog.Info("transaction rollback")
	}
	return
}
