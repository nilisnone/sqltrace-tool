package main

import (
	"sqltrace-go-tool/app"
	"sqltrace-go-tool/tools"
)

func main() {
	tools.InitConfig()
	app := app.NewApplication(tools.GetConfig())
	app.Run()
}
