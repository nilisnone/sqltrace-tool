package tools

import (
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	InitPositionFile string `env:"INIT_POSITION_FILE" envDefault:"/etc/sqltool.ini"` //文件扫描位置记录
	ScanDir          string `env:"SCAN_DIR"`                                         //分析日志的文件路径
	DbDSN            string `env:"DB_DSN"`                                           //数据库连接地址
	ExplainInterval  string `env:"EXPLAIN_TNTERVAL" envDefault:"120"`                //每个SQL指纹执行Explain的间隔
}

var (
	config = &Config{}
)

func InitConfig() {
	_ = godotenv.Load(".sqltool.env.develop")
	_ = godotenv.Load(".sqltool.env")
	_ = godotenv.Load("/etc/.sqltool.env")
	env.Parse(config)
}

func GetConfig() Config {
	return *config
}
