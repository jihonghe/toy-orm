package miniorm

import (
	"database/sql"
	"fmt"
	"strings"

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

// Migrate only supports the column`s add and delete
func (e *Engine) Migrate(value interface{}) (err error) {
	_, err = e.Transaction(func(s *session.Session) (result interface{}, err error) {
		// if table not exist, then try to create it
		exists, err := s.Model(value).TableExists()
		if err != nil {
			return
		}
		if !exists {
			return nil, s.CreateTable()
		}

		// if table exist, then try to migrate it
		table, err := s.RefTable()
		if err != nil {
			return
		}
		rows, err := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		if err != nil {
			return
		}
		oldFields, err := rows.Columns()
		if err != nil {
			return
		}
		// get new columns and columns to be deleted
		newFields := difference(table.FieldNames, oldFields)
		deletedFields := difference(oldFields, table.FieldNames)
		ormlog.Infof("table '%s' migrate: new cols %v, deleted cols %v", table.Name, newFields, deletedFields)
		// add the new fields
		for _, field := range newFields {
			f := table.GetField(field)
			alterSql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s %s", table.Name, f.Name, f.Type, f.Constraints)
			if _, err = s.Raw(alterSql).Exec(); err != nil {
				return
			}
		}
		if len(deletedFields) == 0 {
			return
		}
		// migrate by sql
		tmpTable := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s FROM %s;", tmpTable, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", tmpTable, table.Name))
		_, err = s.Exec()
		return
	})
	return
}

// difference returns element in array a and not in array b.In short, it returns (a - b)
func difference(a, b []string) (diff []string) {
	mapB := make(map[string]struct{})
	for _, e := range b {
		mapB[e] = struct{}{}
	}
	for _, e := range a {
		if _, ok := mapB[e]; !ok {
			diff = append(diff, e)
		}
	}
	return
}
