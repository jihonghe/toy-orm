package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"miniorm"
	"miniorm/ormlog"
)

const (
	tableUser = `
	CREATE TABLE tb_test (
		id int NOT NULL PRIMARY KEY AUTO_INCREMENT,
		name varchar(64) NOT NULL DEFAULT '' UNIQUE,
		age int NOT NULL DEFAULT 0
	)`
)

func main() {
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?loc=Local&parseTime=true", "moresec", "H@CfMSj1qd@9",
		"127.0.0.1", 3307, "db_honeypot")
	engine, err := miniorm.NewEngine("mysql", mysqlDSN)
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	s := engine.NewSession()
	res, err := s.Raw(tableUser).Exec()
	if err != nil {
		ormlog.Error(err)
	}
	rowsAffected, _ := res.RowsAffected()
	ormlog.Infof("rows affected: %d", rowsAffected)
}
