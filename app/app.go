package app

import (
	"os"
	"os/signal"
	"runtime"
	"sqltrace-go-tool/tools"
	"syscall"
	"time"

	"github.com/pyroscope-io/client/pyroscope"
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
	go app.StartPyroscope()

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

func (app *Application) StartPyroscope() {
	if len(app.config.PyroscoeAddr) == 0 {
		return
	}
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	pyroscope.Start(pyroscope.Config{
		ApplicationName: "sqltrace-tool",

		// replace this with the address of pyroscope server
		ServerAddress: app.config.PyroscoeAddr,

		// you can disable logging by setting this to nil
		// Logger: pyroscope.StandardLogger,

		// optionally, if authentication is enabled, specify the API key:
		AuthToken: app.config.PyroscoeToken,

		// you can provide static tags via a map:
		Tags: map[string]string{"hostname": os.Getenv("HOSTNAME")},

		ProfileTypes: []pyroscope.ProfileType{
			// these profile types are enabled by default:
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			// these profile types are optional:
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
}
