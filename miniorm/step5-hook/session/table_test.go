package session

import (
	"testing"
)

var (
	session *Session
)

func init() {
	session = NewSession("sqlite3").Model(&User{})
}

type User struct {
	Id   int    `miniorm:"NOT NULL PRIMARY KEY AUTOINCREMENT"`
	Name string `miniorm:"NOT NULL UNIQUE"`
	Age  int    // no tag miniorm means no constraints
}

func Test_CreateTable(t *testing.T) {
	err := session.CreateTable()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_DropTable(t *testing.T) {
	err := session.DropTable()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_TableExists(t *testing.T) {
	exists, err := session.TableExists()
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatalf("table '%s' is not exist", session.RefTableName())
	}
}
