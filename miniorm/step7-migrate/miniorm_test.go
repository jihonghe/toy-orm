package miniorm

import (
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"miniorm/session"
)

func openDB(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("sqlite3", "./gee.db")
	if err != nil {
		t.Fatal("failed to connect ", err)
	}
	return engine
}

func Test_NewEngine(t *testing.T) {
	engine := openDB(t)
	defer engine.Close()
}

func TestEngine_Transaction(t *testing.T) {
	t.Run("rollback", func(t *testing.T) {
		transactionRollback(t)
	})
	t.Run("commit", func(t *testing.T) {
		transactionCommit(t)
	})
}

func TestEngine_Migrate(t *testing.T) {
	// t.Run("commit", func(t *testing.T) {
	// 	transactionCommit(t)
	// })
	t.Run("migrate", func(t *testing.T) {
		transactionMigrate(t)
	})
}

type User struct {
	Name     string `miniorm:"PRIMARY KEY"`
	Age      int    `miniorm:"NOT NULL DEFAULT 0"`
	HomeAddr string
	Grade    int `miniorm:"NOT NULL DEFAULT 0"`
}

func transactionMigrate(t *testing.T) {
	engine := openDB(t)
	defer engine.Close()
	err := engine.Migrate(&User{})
	if err != nil {
		t.Fatal(err)
	}
}

func transactionCommit(t *testing.T) {
	engine := openDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_, err := engine.Transaction(func(dbSession *session.Session) (result interface{}, err error) {
		_ = dbSession.Model(&User{}).CreateTable()
		_, err = dbSession.Insert(&User{Name: "Tom", Age: 12})
		return
	})
	if err != nil {
		t.Fatal("failed to commit")
	}
	u := &User{}
	err = s.First(u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			t.Fatal("failed to commit")
		}
		t.Fatal("failed to get User info, err: ", err)
	}
}

func transactionRollback(t *testing.T) {
	engine := openDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_, err := engine.Transaction(func(dbSession *session.Session) (result interface{}, err error) {
		_ = dbSession.Model(&User{}).CreateTable()
		_, err = dbSession.Insert(&User{Name: "Tom", Age: 11})
		return nil, errors.New("fake error")
	})
	if err == nil {
		t.Fatal("failed to rollback")
	} else {
		t.Log("rollback, got an err from transaction, err: ", err)
		exists, err := s.TableExists()
		if err != nil {
			t.Fatal("failed to check if table exists, err: ", err)
		}
		if exists {
			t.Fatal("failed to rollback, the table User is still exist")
		}
	}
}
