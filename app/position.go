package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sqltrace-go-tool/tools"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Position struct {
	mu   sync.RWMutex
	data map[string]int
}

func (p *Position) Set(file string, pos int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data[file] = pos
}

func (p *Position) Get(file string) (pos int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data[file]
}

func (p *Position) GetAll() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data
}

func (p *Position) Exist(file string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.data[file]
	return ok
}

func (p *Position) Del(file string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	delete(p.data, file)
}

// initPositionFile 初始化文件 INIT_POSITION_FILE 不存在则创建
func initPositionFile(file string, taskList *TaskList) {
	fd, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = fd.Close(); err != nil {
			tools.LogE("InitPositionFile close file error %s", err.Error())
		}
	}()

	reader := bufio.NewReader(fd)
	for {
		str, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		strS := strings.Split(str, "=")
		if len(strS) == 0 {
			break
		}
		pos := 0
		if len(strS) > 1 {
			pos, err = strconv.Atoi(strings.Trim(strS[1], "\n"))
			if err != nil {
				pos = 0
			}
		}
		taskList.pos.Set(strS[0], pos)
	}
}

// refreshPositionFile 刷新位置文件内容
func refreshPositionFile(posFile string, taskList *TaskList) {
	t := time.NewTicker(time.Second * 1)
	for {
		<-t.C
		t.Reset(time.Second * 600)
		tools.LogD("----RefreshPositionFile----")
		doRefreshPositionFile(posFile, taskList)
	}
}

func doRefreshPositionFile(posFile string, taskList *TaskList) {
	for file := range taskList.GetAll() {
		if !taskList.pos.Exist(file) {
			taskList.pos.Set(file, 0)
		}
	}
	for file := range taskList.pos.GetAll() {
		if !taskList.Exist(file) {
			taskList.pos.Del(file)
		}
	}
	fd, err := os.OpenFile(posFile, os.O_TRUNC|os.O_RDWR|os.O_APPEND, 0700)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = fd.Close(); err != nil {
			tools.LogE("doRefreshPositionFile close file error %s", err.Error())
		}
	}()

	w := bufio.NewWriter(fd)
	for file, pos := range taskList.pos.GetAll() {
		tools.LogD("file is %s, pos is %d", file, pos)
		w.WriteString(fmt.Sprintf("%s=%d\n", file, pos))
	}
	w.Flush()
}

func IsDir(name string) bool {
	f, err := os.Stat(name)
	if err != nil {
		return false
	}
	return f.IsDir()
}
