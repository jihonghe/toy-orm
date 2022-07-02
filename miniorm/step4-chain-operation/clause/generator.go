package clause

import (
	"fmt"
	"strings"
)

// implement the sql clause like INSERT, SELECT and so on

type generator func(values ...interface{}) (sqlClause string, sqlVars []interface{})

var generators map[ClauseType]generator

func init() {
	generators = make(map[ClauseType]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderby
}

// _insert build insert clause like "INSERT INTO tb_test (Name string)"
//  param1: values[0] string, table name
//  param2: values[1] []string, columns
func _insert(values ...interface{}) (sqlClause string, sqlVars []interface{}) {
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields), []interface{}{}
}

// _values build values clause like "VALUES (?), (?)"
//  param values [][]interface{}, means a couple of values
func _values(values ...interface{}) (sqlClause string, sqlVars []interface{}) {
	var builder strings.Builder
	var placeholder string // e.g. "?, ?, ?"

	builder.WriteString("VALUES ")
	for i, value := range values {
		v := value.([]interface{})
		placeholder = genPlaceholders(len(v))
		builder.WriteString(fmt.Sprintf("(%v)", placeholder))
		if i+1 != len(values) {
			builder.WriteString(", ")
		}
		sqlVars = append(sqlVars, v...)
	}
	sqlClause = builder.String()

	return
}

// _select build select clause like "SELECT Name FROM tb_test"
//  param1: values[0] string, table name
//  param2: values[1] []string, fields of selected columns
func _select(values ...interface{}) (sqlClause string, sqlVars []interface{}) {
	tableName := values[0].(string)
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}

// _limit build limit clause like "LIMIT ?"
//  param: value string, the limit number of query result
func _limit(values ...interface{}) (sqlClause string, sqlVars []interface{}) {
	return "LIMIT ?", values
}

// _where build where clause like "WHERE Name = ?"
//  param1: condition-desc like 'NAME = ?"
//  param2: the value of column in condition-desc
func _where(values ...interface{}) (sqlClause string, sqlVars []interface{}) {
	return fmt.Sprintf("WHERE %s", values[0]), values[1:]
}

// _orderby build orderby clause like "ORDER BY Name ASC"
//  param: order-desc string like "Name ASC"
func _orderby(values ...interface{}) (sqlClause string, sqlVars []interface{}) {
	return fmt.Sprintf("ORDER BY %s", values[0]), values[1:]
}

// genPlaceholders build like "?, ?"
func genPlaceholders(num int) (res string) {
	var placeholders []string
	for i := 0; i < num; i++ {
		placeholders = append(placeholders, "?")
	}
	return strings.Join(placeholders, ", ")
}
