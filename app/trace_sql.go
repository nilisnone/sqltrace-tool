package app

import (
	"encoding/json"
)

type TraceSql struct {
	App_uuid          string  // App 请求唯一值
	Sql_uuid          string  // SQL 唯一值
	Trace_sql         string  // 不带参数绑定的 SQL
	Db_host           string  // SQL 运行的机器 IP
	Run_ms            float32 // 执行时间 ms
	Call_sql_position string  // SQL 所在代码位置
	Biz_created_at    string  // SQL 执行时间 2022-12-08T17:32:41.12345
}

func newTraceSql(sql string) (traceSql *TraceSql, ok bool) {
	t := &TraceSql{}
	err := json.Unmarshal([]byte(sql), t)
	if err != nil {
		return t, false
	}
	return t, true
}
