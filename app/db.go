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

	err = db.Ping()
	if err != nil {
		panic(fmt.Sprintf("无法连接数据库: %s", err.Error()))
	}

	db.SetMaxOpenConns(5)
	tools.LogI("设置最大数据库连接数: 5")
	return db
}

func explainSql(db *sql.DB, sql string) (records []map[string]string, err error) {
	rows, err := db.Query("explain " + sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	rc, err := scanRecords(cols, rows)
	if err != nil {
		return nil, err
	}
	return rc, err
}

// 解析记录集
func scanRecords(cols []string, rows *sql.Rows) (records []map[string]string, err error) {
	// init cols
	vals := make([]sql.RawBytes, len(cols)) //sql.RawBytes存储每一单元的值，[]sql.RawBytes存储每行的值
	dest := make([]interface{}, len(cols))  //dest为一行值，每个元素对应数据库行项的slice的指针
	for i := 0; i < len(cols); i++ {
		dest[i] = &vals[i]
	}

	// loop
	for rows.Next() {
		// scan row to dest
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}

		// row vals
		row := map[string]string{}
		for k, v := range vals {
			sv := string(v)
			row[cols[k]] = sv
		}
		records = append(records, row)
	}

	return records, nil
}
