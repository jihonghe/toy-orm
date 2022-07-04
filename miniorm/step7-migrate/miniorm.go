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

type TxFunc func(*session.Session) (interface{}, error)

/*
Transaction is a convenient method to do a transaction.
	Param f TxFunc, it is a callback function and includes all DB operation that makes up a transaction.
	An example of TxFunc:
	func RecreateAndInsert(s *session.Session) (result interface{}, err error) {
		s.Model(&User{}).DropTable()
		s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{Id: 1, Name: "Tom", Age: 18})
		return
	}

	And then, we can use RecreateAndInsert as the callback func of Transaction
*/
func (e *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := e.NewSession()
	if err = s.Begin(); err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p) // re-throw the panic after rollback
		} else if err != nil {
			_ = s.Rollback() // err is not nil, just rollback
		} else {
			err = s.Commit() // err is nil, commit and update err
		}
	}()

	return f(s)
}
