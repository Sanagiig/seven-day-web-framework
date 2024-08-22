package main

import (
	"fmt"
	"seven-day-web-framework/geeorm"
)

func main() {
	engine, _ := geeorm.NewEngine("mysql", "root:lwj@1993@tcp(127.0.0.1:3306)/zhihu_go")
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	result, _ := s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	count, _ := result.RowsAffected()
	fmt.Printf("Exec success, %d affected\n", count)
}
