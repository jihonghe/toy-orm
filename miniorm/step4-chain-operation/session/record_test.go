package session

import (
	"testing"
)

var (
	u1 = &User{1, "Tom", 10}
	u2 = &User{2, "Sam", 11}
	u3 = &User{3, "Jerry", 12}
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
