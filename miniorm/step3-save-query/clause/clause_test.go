package clause

import (
	"reflect"
	"testing"
)

func TestSelect(t *testing.T) {
	var clause Clause
	clause.Set(SELECT, "User", []string{"*"})
	clause.Set(WHERE, "Name = ?", "Tom")
	clause.Set(ORDERBY, "Name ASC")
	clause.Set(LIMIT, 3)
	sqlClause, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	t.Log(sqlClause, ",", vars)

	expectedSql := "SELECT * FROM User WHERE Name = ? ORDER BY Name ASC LIMIT ?"
	if sqlClause != expectedSql {
		t.Fatalf("failed to generate sql: \nexpected sql: '%s'\nactual sql: '%s'", expectedSql, sqlClause)
	}
	expectedVars := []interface{}{"Tom", 3}
	if !reflect.DeepEqual(vars, expectedVars) {
		t.Fatalf("failed to generate expected sql vars %v, actual vars: %v", expectedVars, vars)
	}
}

func TestClause_Build(t *testing.T) {
	var clause Clause
	clause.Set(INSERT, "User")

}
