package main

import (
	"sqltrace-go-tool/app"
	"sqltrace-go-tool/tools"
)

func main() {
	app.NewApplication(tools.GetConfig()).Run()
}
