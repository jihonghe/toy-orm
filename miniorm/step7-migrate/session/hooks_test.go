package session

import (
	"testing"

	"miniorm/ormlog"
)

func TestUserHook(t *testing.T) {
	s := testRecord(t)
	var users []User
	err := s.Find(&users)
	if err != nil {
		t.Fatalf("failed to get users, err: %v", err)
	}
	ormlog.Info(users)
}
