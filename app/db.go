package app

import (
	"database/sql"
	"fmt"
	"sqltrace-go-tool/tools"

	_ "github.com/go-sql-driver/mysql"
)

func initDB(dsn string) *sql.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("无法设置数据库，配置: %s", err.Error()))
	}
	defer db.Close()
	// rows, err := db.Query("explain FORMAT = JSON select runoob_title from runoob_tbl")
	// if err != nil {
	// 	logE("error %s", err)
	// 	return
	// }
	// defer rows.Close()
	// for rows.Next() {
	// 	var s string
	// 	err := rows.Scan(&s)
	// 	if err != nil {
	// 	}
	// 	fmt.Println(s)
	// }

	err = db.Ping()
	if err != nil {
		panic(fmt.Sprintf("无法连接数据库: %s", err.Error()))
	}

	db.SetMaxOpenConns(5)
	tools.LogI("设置最大数据库连接数: 5")
	return db
}
