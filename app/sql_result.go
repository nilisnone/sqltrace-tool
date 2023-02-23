package app

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slices"
)

type ResultMap struct {
	mu   sync.RWMutex
	data map[string]*SqlStatistic
}

func (rm *ResultMap) Set(finger_md5 string, ss *SqlStatistic) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.data[finger_md5] = ss
}

func (rm *ResultMap) Exist(finger_md5 string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	_, ok := rm.data[finger_md5]
	return ok
}

func (rm *ResultMap) Get(finger_md5 string) *SqlStatistic {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.data[finger_md5]
}

func (rm *ResultMap) SetNx(finger_md5 string, ss *SqlStatistic) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if _, ok := rm.data[finger_md5]; ok {
		rm.data[finger_md5] = ss
		return true
	}
	return false
}

func (rm *ResultMap) GetAllWithNoLock() map[string]*SqlStatistic {
	return rm.data
}

type SqlStatistic struct {
	mu sync.RWMutex

	Finger     string // SQL 指纹
	Finger_md5 string // SQL 指纹 md5

	// Explain 信息
	Explain_result        string // explain 的原始结果
	Last_explain_time     string // 最近 explain 的时间
	Last_explain_timestap int64  // 最近 explain 的时间戳

	// 分析信息
	Is_all      bool // 是否全表扫描
	Is_filesort bool // 是否用到文件排序
	Is_temp     bool // 是否用到临时表

	// 全局统计相关信息
	Arise_total_times int            // 出现次数
	Duplicate_times   int            // 同一个 app_uuid 中，多次出现当前 SQL 的最大次数
	App_uuid_times    map[string]int // 同一个 app_uuid 中，多次出现当前 SQL 的次数
	Appear_files      []string       // 出现当前 SQL 的所有可能位置
}

func NewSqlStatisticFromTraceSql(finger string, finger_md5 string, traceSql *TraceSql) *SqlStatistic {
	return &SqlStatistic{
		Finger:     finger,
		Finger_md5: finger_md5,

		Arise_total_times: 1,
		Duplicate_times:   1,
		App_uuid_times: map[string]int{
			traceSql.App_uuid: 1,
		},
		Appear_files: []string{
			traceSql.Call_sql_position,
		},
	}
}

// SetExplainResult 设置 explain 结果信息
func (ss *SqlStatistic) SetExplainResult(records []map[string]string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	res, _ := json.Marshal(records)
	ss.Explain_result = string(res)
	ss.Last_explain_time = time.Now().Local().Format(time.RFC3339)
	ss.Last_explain_timestap = time.Now().Unix()
}

// IssueExplore 发现 SQL 的问题
func (ss *SqlStatistic) IssueExplore(records []map[string]string) {
	isAll := false
	isFileSort := false
	isTemp := false
	// Using where; Using temporary; Using filesort
	for _, record := range records {
		for k, v := range record {
			if k == "type" && !isAll {
				isAll = v == "ALL"
			}
			if k == "Extra" {
				if !isFileSort {
					isFileSort = strings.Contains(v, "Using filesort")
				}
				if !isTemp {
					isTemp = strings.Contains(v, "Using temporary")
				}
			}
		}
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.Is_all = isAll
	ss.Is_filesort = isFileSort
	ss.Is_temp = isTemp
}

// GlobalStatistics 全局统计信息
func (ss *SqlStatistic) GlobalStatistics(traceSql *TraceSql, exist bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	// 如果存在
	if exist {
		ss.Arise_total_times++ // 累计出现次数+1
		ss.App_uuid_times[traceSql.App_uuid]++
		//更新最大值 Duplicate_times
		if ss.App_uuid_times[traceSql.App_uuid] > ss.Duplicate_times {
			ss.Duplicate_times = ss.App_uuid_times[traceSql.App_uuid]
		}
		if !slices.Contains(ss.Appear_files, traceSql.Call_sql_position) {
			ss.Appear_files = append(ss.Appear_files, traceSql.Call_sql_position)
		}
	}
}
