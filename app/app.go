package app

import (
	"os"
	"os/signal"
	"sqltrace-go-tool/tools"
	"syscall"
	"time"
)

type Application struct {
	config   tools.Config
	taskList *TaskList
}

func NewApplication(config tools.Config) *Application {
	app := &Application{}

	app.config = config
	app.taskList = NewTaskList()
	return app
}

func (app *Application) Run() {
	tools.CheckConfig()
	app.setDB()
	defer app.taskList.db.Close()

	tools.LogI("开始初始化扫描文件起始行, 初始化文件 %s", app.config.InitPositionFile)
	initPositionFile(app.config.InitPositionFile, app.taskList)

	tools.LogI("开始加载文件夹 %s 到任务列表", app.config.ScanDir)
	initTaskListFromLogFile(app.config.ScanDir, app.taskList)

	tools.LogI("开始监控文件夹 %s 的新文件", app.config.ScanDir)
	go watchNewLogFile(app.config.ScanDir, app.taskList, extraceTraceSqlFromFile)

	tools.LogI("开始定时刷新扫描文件起始行，到初始化文件中 %s", app.config.InitPositionFile)
	go refreshPositionFile(app.config.InitPositionFile, app.taskList)

	go app.startTask()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	c := <-quit
	tools.LogI("接收到信号 %s[%d]", c.String(), c)
	doRefreshPositionFile(app.config.InitPositionFile, app.taskList)
}

func (app *Application) setDB() {
	app.taskList.db = initDB(app.config.DbDSN)
}

// StartTask 开始任务
func (app *Application) startTask() {
	tools.LogI("after 3s ... will start task")
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	go app.saveStatisticResult()

	for {
		<-t.C
		for file, task := range app.taskList.GetAll() {
			tools.LogI("file[%s] start task", file)
			if app.taskList.pos.Exist(file) {
				go extraceTraceSqlFromFile(file, task.fd, app.taskList.pos.Get(file), app.taskList)
			}
		}
		t.Reset(time.Second * 300)
	}
}

// saveStatisticResult 异步保存分析结果到文件中
func (app *Application) saveStatisticResult() {
	t := time.NewTimer(time.Second * 300)

	for {
		<-t.C
		t.Reset(time.Second * 300)
		stLog(app.taskList.result.GetAllWithNoLock())
	}
}
