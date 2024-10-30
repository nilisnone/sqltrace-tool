package tools

import (
	"encoding/json"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	InitPositionFile string `env:"INIT_POSITION_FILE" envDefault:"/etc/sqltool.ini"` //文件扫描位置记录
	ScanDir          string `env:"SCAN_DIR"`                                         //分析日志的文件路径
	DbDSN            string `env:"DB_DSN"`                                           //数据库连接地址
	ExplainInterval  int    `env:"EXPLAIN_TNTERVAL" envDefault:"120"`                //每个SQL指纹执行Explain的间隔
	SinkDriver       string `env:"SINK_DRIVER" envDefault:"file"`
	SinkFileDir      string `env:"SINK_FILE_DIR" envDefault:""` // 保存结果文件夹, 如果SinkDriver是file，默认为 SCAN_DIR
	DbDSNMap         map[string]map[string]string
}

var (
	config = &Config{}
)

func init() {
	_ = godotenv.Load(".sqltool.env.develop")
	_ = godotenv.Load(".sqltool.env")
	_ = godotenv.Load("/etc/.sqltool.env")
	env.Parse(config)
	parseDSN(config)
	if config.SinkDriver == "file" && len(config.SinkFileDir) < 1 {
		config.SinkFileDir = config.ScanDir
	}
}

func parseDSN(config *Config) {
	// 解析 JSON 字符串
	err := json.Unmarshal([]byte(config.DbDSN), &config.DbDSNMap)
	if err != nil {
		panic(err)
	}
	// fmt.Println(config.DbDSNMap)
}

func CheckConfig() {
	config := GetConfig()
	if len(config.ScanDir) < 1 {
		panic("SCAN_DIR 不能为空")
	}
	if len(config.DbDSN) < 1 {
		panic("DB_DSN 不能为空")
	}
}

func GetConfig() Config {
	return *config
}
