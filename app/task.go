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
	records, err := explainSql(taskList.db, traceSql.Trace_sql)
	if err != nil {
		tools.LogE("explain 失败: %s", err.Error())
		return false
	}
	sqlStatistic.SetExplainResult(records)
	sqlStatistic.IssueExplore(records)
	return true
}

// afterExecSqlExplaining 执行 explain 之后动作
func afterExecSqlExplaining(sqlStatistic *SqlStatistic, taskList *TaskList) {
	j, _ := json.Marshal(sqlStatistic)
	tools.LogD("%s", string(j))

	taskList.result.Set(sqlStatistic.Finger_md5, sqlStatistic)
}
