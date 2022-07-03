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
	INSERT  ClauseType = iota // param1: tableName string; param2: columns ...interface{}
	VALUES                    // param: values [][]interface{}, means a couple of values
	SELECT                    // param1: tableName string; param2: columns ...interface{}
	LIMIT                     // param1: offset uint64; param2: limit uint64
	WHERE                     // param1: conditionDesc string; param2: values ...interface{}
	ORDERBY                   // param: order-desc string like "Name ASC"
	UPDATE                    // param: the fields and new values, supports map[string]interface{} and (field, value)
	DELETE                    // param: tableName string
	COUNT                     // param: tableName string
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
