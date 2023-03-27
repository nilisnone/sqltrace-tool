package app

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"sqltrace-go-tool/tools"
	"time"

	"github.com/percona/go-mysql/query"
)

type Task struct {
	file string
	fd   *os.File
}

// parseSql 解析SQL
func parseSql(traceSql *TraceSql, taskList *TaskList) {
	sqlStatistic, needExpain := beforeExecSqlExplaining(traceSql, taskList)
	defer saveRealtimeResult(traceSql, sqlStatistic, taskList)
	if !needExpain {
		tools.LogD("finger_md5 = %s, exist and not need explain again", sqlStatistic.Finger_md5)
		taskList.result.Set(sqlStatistic.Finger_md5, sqlStatistic)
		return
	}
	if ok := execSqlExplaining(traceSql, sqlStatistic, taskList); !ok {
		return
	}
	afterExecSqlExplaining(sqlStatistic, taskList)
}

// beforeExecSqlExplaining 执行 explain 之前动作
func beforeExecSqlExplaining(traceSql *TraceSql, taskList *TaskList) (SqlStatistic *SqlStatistic, needExpain bool) {
	finger := query.Fingerprint(traceSql.Trace_sql)
	tools.LogI("finger = %s ", finger)
	fingerMd5 := fmt.Sprintf("%x", md5.Sum([]byte(finger)))

	ss := NewSqlStatisticFromTraceSql(finger, fingerMd5, traceSql)
	exist := false
	if taskList.result.Exist(fingerMd5) {
		ss = taskList.result.Get(fingerMd5)
		exist = true
	}
	ss.GlobalStatistics(traceSql, exist)
	if !exist { // 需要 explain(1): 第一次
		return ss, true
	}
	// 需要 explain(2): 最近一次 explain 时间超过 {EXPLAIN_TNTERVAL} 秒
	if time.Now().Local().Unix() > ss.Last_explain_timestap+int64(tools.GetConfig().ExplainInterval) {
		return ss, true
	}

	return ss, false
}

// execSqlExplaining 执行 explain
func execSqlExplaining(traceSql *TraceSql, sqlStatistic *SqlStatistic, taskList *TaskList) (ok bool) {
	db, ok := taskList.db[traceSql.Db_alias]
	if !ok {
		tools.LogW("%s 没有配置对应的 DB_DSN", traceSql.Db_alias)
		return false
	}
	driver := GetExplainer(db.Driver)
	records, err := driver.explainSql(db.DSN, traceSql.Trace_sql)
	if err != nil {
		tools.LogE("explain 失败: %s", err.Error())
		return false
	}
	sqlStatistic.SetExplainResult(records)
	driver.issueExplore(sqlStatistic, records)
	return true
}

// afterExecSqlExplaining 执行 explain 之后动作
func afterExecSqlExplaining(sqlStatistic *SqlStatistic, taskList *TaskList) {
	j, _ := json.Marshal(sqlStatistic)
	tools.LogD("%s", string(j))

	taskList.result.Set(sqlStatistic.Finger_md5, sqlStatistic)
}

type rt struct {
	App_uuid  string // App 请求唯一值
	Sql_uuid  string // SQL 唯一值
	Trace_sql string // 不带参数绑定的 SQL

	Finger     string // SQL 指纹
	Finger_md5 string // SQL 指纹 md5

	// Explain 信息
	Explain_result    string // explain 的原始结果
	Last_explain_time string // 最近 explain 的时间

	// 分析信息
	Is_all      bool // 是否全表扫描
	Is_filesort bool // 是否用到文件排序
	Is_temp     bool // 是否用到临时表

}

// saveRealtimeResult 保存实时分析结果到文件中
func saveRealtimeResult(traceSql *TraceSql, sqlStatistic *SqlStatistic, taskList *TaskList) {
	r := &rt{
		App_uuid:  traceSql.App_uuid,
		Sql_uuid:  traceSql.Sql_uuid,
		Trace_sql: traceSql.Trace_sql,

		Finger:     sqlStatistic.Finger,
		Finger_md5: sqlStatistic.Finger_md5,

		Explain_result:    sqlStatistic.Explain_result,
		Last_explain_time: sqlStatistic.Last_explain_time,

		Is_all:      sqlStatistic.Is_all,
		Is_filesort: sqlStatistic.Is_filesort,
		Is_temp:     sqlStatistic.Is_temp,
	}
	j, _ := json.Marshal(r)
	rtLog(string(j))
}
