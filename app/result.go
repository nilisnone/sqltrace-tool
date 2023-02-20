package app

import "sync"

type ResultMap struct {
	mu   sync.RWMutex
	data map[string]SqlStatic
}

type SqlStatic struct {
	Finger_md5            string // SQL指纹md5
	Arise_times           int    // 出现次数
	Last_expain_timestamp int    // 最近explain的时间戳
}

func NewSqlStatic(finger string, t int, ts int) *SqlStatic {
	return &SqlStatic{Finger_md5: finger, Arise_times: t, Last_expain_timestamp: ts}
}

func NewResultMap() *ResultMap {
	return &ResultMap{
		data: make(map[string]SqlStatic),
	}
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
