package app

import (
	"bufio"
	"encoding/json"
	"os"
	"sqltrace-go-tool/tools"
	"strings"
	"time"
)

var sinkFileRt *os.File
var sinkFileRtW *bufio.Writer

func newSinkFileRt() (sf *bufio.Writer, err error) {
	if sinkFileRt != nil {
		if strings.HasSuffix(sinkFileRt.Name(), time.Now().Local().Format("20060102")) {
			return sinkFileRtW, nil
		}
		sinkFileRt.Close()
		sinkFileRt = nil
	}

	config := tools.GetConfig()
	sinkFileRt, err := os.OpenFile(config.SinkFileDir+"/realtime.log."+time.Now().Local().Format("20060102"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	sinkFileRtW = bufio.NewWriter(sinkFileRt)

	return sinkFileRtW, nil
}

func rtLog(msg string) {
	sf, err := newSinkFileRt()
	if err != nil {
		tools.LogE("写实时日志失败: %s", err.Error())
	}
	sf.WriteString(msg + "\n")
	sf.Flush()
}

var sinkFileSt *os.File
var sinkFileStW *bufio.Writer

func newSinkFileSt() (sf *bufio.Writer, err error) {
	if sinkFileSt != nil {
		if strings.HasSuffix(sinkFileSt.Name(), time.Now().Local().Format("20060102")) {
			return sinkFileStW, nil
		}
		sinkFileSt.Close()
		sinkFileSt = nil
	}

	config := tools.GetConfig()
	sinkFileSt, err := os.OpenFile(config.SinkFileDir+"/statistic.log."+time.Now().Local().Format("20060102"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	sinkFileStW = bufio.NewWriter(sinkFileSt)

	return sinkFileStW, nil
}

func stLog(ss map[string]*SqlStatistic) {
	sf, err := newSinkFileSt()
	if err != nil {
		tools.LogE("写统计日志失败: %s", err.Error())
		return
	}
	for _, s := range ss {
		f, _ := json.Marshal(s)
		sf.WriteString(string(f) + "\n")
	}

	sf.Flush()
}
