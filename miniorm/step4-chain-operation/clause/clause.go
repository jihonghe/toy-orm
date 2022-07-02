package clause

import (
	"strings"
)

type Clause struct {
	sql     map[ClauseType]string
	sqlVars map[ClauseType][]interface{}
}

type ClauseType int

const (
	INSERT ClauseType = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
)

// Set gen sql clause based on the given clause type and vars, and then save it in Clause instance
//  param name ClauseType, the sql clause(insert,where and so on)
//  param vars[0] string, table name
//  param vars[1:] []string, columns(included column constraints) of table
func (c *Clause) Set(name ClauseType, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[ClauseType]string)
		c.sqlVars = make(map[ClauseType][]interface{})
	}
	sqlClause, vars := generators[name](vars...)
	c.sql[name] = sqlClause
	c.sqlVars[name] = vars
}

// Build generate the complete sql based the given clause order
func (c *Clause) Build(orders ...ClauseType) (sqlClause string, vars []interface{}) {
	var clauses []string
	var clauseVars []interface{}
	for _, order := range orders {
		if sqlC, ok := c.sql[order]; ok {
			clauses = append(clauses, sqlC)
			clauseVars = append(clauseVars, c.sqlVars[order]...)
		}
	}
	return strings.Join(clauses, " "), clauseVars
}
