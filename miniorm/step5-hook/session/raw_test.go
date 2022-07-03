package session

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"miniorm/dialect"
)

var (
	TestDB *sql.DB
)

func Init() {
	var err error
	if TestDB == nil {
		TestDB, err = sql.Open("sqlite3", "../gee.db")
		if err != nil {
			panic(err)
		}
	}
}

func NewSession(dbType string) *Session {
	Init()
	dial, ok := dialect.GetDialect(dbType)
	if !ok {
		panic(dbType + " is not registered")
	}
	return New(TestDB, dial)
}

func TestSession_Exec(t *testing.T) {
	s := NewSession("sqlite3")
	s.Raw("DROP TABLE IF EXISTS User;").Exec()
	s.Raw("CREATE TABLE User(Name text, Age integer);").Exec()
	res, _ := s.Raw("INSERT INTO User(`Name`, `Age`) values (?,?), (?,?);", "Tom", 3, "Sam", 5).Exec()
	affected, err := res.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}
	if affected != 2 {
		t.Fatal("expect 2, but get ", affected)
	}
}
