package miniorm

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
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
