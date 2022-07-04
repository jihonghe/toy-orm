package session

import (
	"testing"

	"miniorm/ormlog"
)

var (
	session *Session
)

func init() {
	session = NewSession("sqlite3").Model(&User{})
}

type User struct {
	Id            int    `miniorm:"NOT NULL PRIMARY KEY AUTOINCREMENT"`
	Name          string `miniorm:"NOT NULL UNIQUE"`
	Age           int    // no tag miniorm means no constraints
	PrivateSecret string
}

func (u *User) AfterQuery(s *Session) (err error) {
	ormlog.Debugf("[hook-call] User.AfterQuery, user: %v", u)
	u.PrivateSecret = "---------------------"
	return
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
