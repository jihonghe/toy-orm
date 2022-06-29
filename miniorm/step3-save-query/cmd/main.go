package main

import (
	_ "github.com/mattn/go-sqlite3"

	"miniorm"
	"miniorm/ormlog"
)

func main() {
	engine, err := miniorm.NewEngine("sqlite3", "../gee.db")
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	s := engine.NewSession()
	s.Raw("DROP TABLE IF EXISTS User;").Exec()
	s.Raw("CREATE TABLE User(Name text, Age integer);").Exec()
	s.Raw("CREATE TABLE User(Name text, Age integer);").Exec()
	res, _ := s.Raw("INSERT INTO User(`Name`, `Age`) values(?,?), (?,?)", "Tom", 5, "Sam", 7).Exec()
	affected, _ := res.RowsAffected()
	ormlog.Infof("rows-affected: %d", affected)
}
