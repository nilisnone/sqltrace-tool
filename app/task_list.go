package app

import (
	"bytes"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"sqltrace-go-tool/tools"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type TaskList struct {
	mu     sync.RWMutex
	data   map[string]*Task
	pos    *Position
	result *ResultMap
	db     map[string]*DsnMap // alias: {driver: mysql, dsn: *sql.DB}
}

type DsnMap struct {
	Driver string
	DSN    *sql.DB
}

func NewTaskList() *TaskList {
	return &TaskList{
		data: make(map[string]*Task),
		pos: &Position{
			data: make(map[string]int),
		},
		result: &ResultMap{
			data: make(map[string]*SqlStatistic),
		},
		db: make(map[string]*DsnMap),
	}
}

func (t *TaskList) Set(file string, ts *Task) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.data[file] = ts
}

func (t *TaskList) Del(file string) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	delete(t.data, file)
}

func (t *TaskList) Get(file string) *Task {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.data[file]
}

func (t *TaskList) GetAll() map[string]*Task {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.data
}

func (t *TaskList) Exist(file string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.data[file]
	return ok
}

// initTaskListFromLogFile 初始化 SCAN_DIR 中的文件，加入到 task 中
func initTaskListFromLogFile(dir string, taskList *TaskList) {
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			base, _ := filepath.Abs(dir)
			fullPath := filepath.Join(base, info.Name())
			if !strings.HasSuffix(fullPath, ".log") {
				tools.LogI("fullpath=%s, %s", fullPath, " is not with .log suffix")
				return nil
			}
			fd, err := os.OpenFile(path, os.O_RDONLY, 0700)
			if err != nil {
				return err
			}
			taskList.Set(fullPath, &Task{fullPath, fd})
			tools.LogD("初始化文件加入任务队列, file = %s", fullPath)
		}
		return nil
	})
	if err != nil {
		tools.LogE("初始化读取文件夹{%s}失败 %s", dir, err.Error())
	}
}

// watchNewLogFile 监控 SCAN_DIR 里的新文件，加入到 task 中
func watchNewLogFile(dir string, taskList *TaskList, callback func(string, *os.File, int, *TaskList)) {
	time.Now().GoString()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		tools.LogE("监控文件错误 %s", err.Error())
	}
	defer func() {
		if err = watcher.Close(); err != nil {
			tools.LogE("WatchNewFile close watcher error %s", err.Error())
		}
	}()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// tools.LogD("监控事件 = %s, 文件 = %s", event.Op, event.Name)
				fullPath := event.Name
				if !IsDir(event.Name) {
					base, _ := filepath.Abs(dir)
					fullPath = filepath.Join(base, filepath.Base(event.Name))
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					if IsDir(event.Name) {
						tools.LogD("%s 是新文件夹,添加监控事件", event.Name)
						err = watcher.Add(event.Name)
						if err != nil {
							tools.LogE("新建新文件夹监控事件失败 %s", err.Error())
						}
					} else {
						if !strings.HasSuffix(fullPath, ".log") {
							tools.LogI("path=%s,%s", fullPath, " is not with .log suffix")
							return
						}
						fd, err := os.OpenFile(event.Name, os.O_RDONLY, 0700)
						if err != nil {
							tools.LogE("打开新文件失败 %s", err.Error())
						} else {
							tools.LogI("新文件加入队列成功, file=%s", fullPath)
							taskList.Set(fullPath, &Task{event.Name, fd})
							go extraceTraceSqlFromFile(fullPath, fd, 0, taskList)
						}
					}
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					if IsDir(event.Name) {
						tools.LogD("%s 是文件夹,移除监控事件", event.Name)
						watcher.Remove(event.Name)
					} else {
						// 如果已经在taskList中则移除
						if taskList.Exist(fullPath) {
							tools.LogD("任务列表中移除 %s", fullPath)
							taskList.Del(fullPath)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				tools.LogE("监控文件错误2 %s", err.Error())
			}
		}
	}()
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				tools.LogE("监控文件夹 %s 错误 %s", path, err.Error())
			} else {
				tools.LogD("开始监控文件夹 %s", path)
			}
		}
		return nil
	})

	done := make(chan bool)
	<-done
}

// extraceTraceSqlFromFile 依次读取文件按行提取任务，每5分钟退出一次
func extraceTraceSqlFromFile(file string, fd *os.File, start int, taskList *TaskList) {
	t := time.NewTicker(time.Second)
	ts := time.NewTimer(time.Second * 300)
	defer func() {
		t.Stop()
		ts.Stop()
	}()
	st := start
	for {
		select {
		case <-t.C:
			for {
				str, next_pos, err := tools.ExtractLine(fd, st)
				if err != nil {
					tools.LogE("遇到错误：%s", err.Error())
				}
				if next_pos == st {
					// 没有数据了，更新task
					taskList.pos.Set(file, next_pos)
					break
				}
				str = string(bytes.Trim([]byte(str), "\x00"))
				tools.LogD("file: %s, start: %d, end: %d, str: `%s`", file, st, next_pos, str)
				st = next_pos
				traceSql, ok := newTraceSql(str)
				if !ok {
					tools.LogW("file: %s, start: %d, end: %d cannot found trace-sql", file, st, next_pos)
					taskList.pos.Set(file, next_pos)
					continue
				}
				tools.LogI("app_uuid: %s, sql_uuid: %s", traceSql.App_uuid, traceSql.Sql_uuid)
				tools.LogD("sql = %s ", traceSql.Trace_sql)
				taskList.pos.Set(file, next_pos)
				parseSql(traceSql, taskList)
			}
		case <-ts.C:
			return
		}
	}
}
