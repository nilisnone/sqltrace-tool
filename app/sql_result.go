package app

import "sync"

type ResultMap struct {
	mu   sync.RWMutex
	data map[string]SqlStatic
}

type SqlStatic struct {
	Finger                string   // SQL 指纹
	Finger_md5            string   // SQL 指纹 md5
	Arise_total_times     int      // 出现次数
	Duplicate_times       int      // 同一个 app_uuid 中，多次出现当前 SQL 的次数
	App_uuids             []string // 出现当前 SQL 的所有 app_uuid
	Explain_result        string   // explain 的原始结果
	Appear_files          []string // 出现当前 SQL 的所有可能位置
	Last_expain_timestamp int      // 最近 explain 的时间戳
	Last_timestamp        int      // 最新出现的时间戳
}

func (rm *ResultMap) Set(finger string, ss *SqlStatic) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.data[finger] = *ss
}

func (rm *ResultMap) Exist(finger string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	_, ok := rm.data[finger]
	return ok
}

type SqlExplain struct {
	Finger        string // SQL 指纹
	Finger_md5    string // SQL 指纹 md5
	App_uuid      string // 生命周期 uuid
	Sql_uuid      string // sql uuid
	Expain_result string // explain 的原始结果
	Is_all        bool   // 是否全表扫描
	Is_filesort   bool   // 是否用到文件排序
	Is_temp       bool   // 是否用到临时表
}
