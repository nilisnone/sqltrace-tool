package tools

import (
	"github.com/caarlos0/env"
	"github.com/go-basic/ipv4"
	"github.com/joho/godotenv"
)

type Config struct {
	InitPositionFile   string `env:"INIT_POSITION_FILE" envDefault:"/etc/sqltool.ini"` //文件扫描位置记录
	ScanDir            string `env:"SCAN_DIR"`                                         //分析日志的文件路径
	DbDSN              string `env:"DB_DSN"`                                           //数据库连接地址
	ExecutorIp         string `env:"XXL_EXECUTOR_IP"`                                  //执行器IP
	ExecutorPort       int    `env:"XXL_EXECUTOR_PORT" envDefault:"9999"`              //执行器端口
	AppName            string `env:"XXL_APP_NAME"`                                     //执行器名称
	PhpCommand         string `env:"XXL_PHP" envDefault:"php"`                         //php项目入口文件,laravel项目对应artisan文件
	PhpIndex           string `env:"XXL_ARTISAN" envDefault:"/var/www/artisan"`        //php执行路径
	LogLevel           string `env:"XXL_LOG_LEVEL" envDefault:"info"`                  //日志级别
	LogDir             string `env:"XXL_LOG_DIR" envDefault:"/tmp/xxl-gogent"`         //日志文件dir
	WaitQuitSecond     int    `env:"XXL_WAIT_QUIT_SECOND" envDefault:"60"`             //等待进程退出的秒数
	DaemonPeriodSecond int    `env:"XXL_DAEMON_PERIOD_SECOND" envDefault:"30"`         //守护进程处理间隔时间秒
	DaemonMaxRestart   int    `env:"XXL_DAEMON_MAX_RESTART" envDefault:"3"`            //守护进程最多重启次数
	MaxPerFileSize     int    `env:"XXL_MAX_PER_FILE_SIZE" envDefault:"20"`            //单个文件最大多少M后清空
	PyroscopeHost      string `env:"XXL_PYROSCOPE_HOST" envDefault:""`                 // pyroscope 地址
	PyroscopeToken     string `env:"XXL_PYROSCOPE_TOKEN" envDefault:""`                // pyroscope token
	PerStopSecond      int    `env:"XXL_PER_STOP_SENCOND" envDefault:"2"`              // 慢启动等待秒，如果系统依赖第三方资源，比如.env文件，可以设置等待多少秒
}

var (
	config = &Config{}
)

func InitConfig() {
	_ = godotenv.Load(".sqltool.env.develop")
	_ = godotenv.Load(".sqltool.env")
	_ = godotenv.Load("/etc/.sqltool.env")
	env.Parse(config)
	if len(config.ExecutorIp) == 0 {
		config.ExecutorIp = ipv4.LocalIP()
	}
}

func GetConfig() Config {
	return *config
}
