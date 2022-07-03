package schema

import (
	"testing"

	"miniorm/dialect"
)

type User struct {
	Id   int    `miniorm:"NOT NULL PRIMARY KEY AUTO_INCREMENT"`
	Name string `miniorm:"NOT NULL UNIQUE"`
	Age  int    // no tag miniorm means no constraints
}

var testDial, _ = dialect.GetDialect("sqlite3")

func TestParse(t *testing.T) {
	schema := Parse(&User{}, testDial)
	if schema.Name != "User" || len(schema.Fields) != 3 {
		t.Fatalf("failed to parse User struct, schema name: %s, fields: %d", schema.Name, len(schema.Fields))
	}
	if schema.GetField("Name").Constraints != "NOT NULL UNIQUE" {
		t.Fatal("failed to parse tag of struct field Name")
	}
}
