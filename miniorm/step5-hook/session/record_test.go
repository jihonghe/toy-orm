package session

import (
	"testing"

	"miniorm/ormlog"
)

var (
	u1 = &User{1, "Tom", 10, "Tom`s private secret"}
	u2 = &User{2, "Sam", 11, "Sam`s private secret"}
	u3 = &User{3, "Jerry", 12, "Jerry`s private secret"}
)

func testRecord(t *testing.T) (s *Session) {
	t.Helper()
	s = NewSession("sqlite3").Model(&User{})
	err1 := s.DropTable()
	err2 := s.CreateTable()
	_, err3 := s.Insert(u1, u3)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("failed to init test records:\ndrop-table-err: %v\ncreate-table-err: %v\ninsert-err: %v",
			err1, err2, err3)
	}

	return
}

func TestSession_First(t *testing.T) {
	s := testRecord(t)
	var u User
	err := s.First(&u)
	if err != nil {
		t.Fatalf("failed to query data, err: %v", err)
	}
	ormlog.Info(u)
}

func TestSession_Count(t *testing.T) {
	s := testRecord(t)
	count, err := s.Where("Age > ?", 10).Count()
	if err != nil {
		t.Fatalf("failed to count records, err: %v", err)
	}
	ormlog.Info(count)
}

func TestSession_Where(t *testing.T) {
	s := testRecord(t)
	var users []User
	err := s.Where("Age = ? or Name like ?", 13, `%T%`).Find(&users)
	if err != nil {
		t.Fatalf("failed to get users, err: %v", err)
	}
	ormlog.Info(users)
}

func TestSession_Update(t *testing.T) {
	t.Log("Before update: ")
	TestSession_Find(t)
	s := testRecord(t)
	// TODO: it will failed to update if UNIQUE in some field`s constraints without the WHERE clause
	rowsAffected, err := s.Update("Age", 13)
	if err != nil {
		t.Fatalf("failed to update, err: %v", err)
	}
	t.Logf("rows affected: %d", rowsAffected)

	t.Log("After update: ")
	var users []User
	err = s.Find(&users)
	if err != nil {
		t.Fatalf("failed to find user records, err: %v", err)
	}
	t.Log(users)
}

func TestSession_Insert(t *testing.T) {
	s := testRecord(t)
	rowsAffected, err := s.Insert(u2)
	if err != nil || rowsAffected != 1 {
		t.Fatalf("failed to insert record: %v, err: %v", u3, err)
	}
}

func TestSession_Find(t *testing.T) {
	s := testRecord(t)
	var users []User
	err := s.Find(&users)
	if err != nil {
		t.Fatalf("failed to find user records, err: %v", err)
	}
	t.Log(users)
}
