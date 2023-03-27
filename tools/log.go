package tools

import (
	"fmt"
	"time"
)

func LogI(format string, a ...interface{}) {
	fmt.Printf("[%s][info] %s\n", time.Now().Local().Format(time.RFC3339), fmt.Sprintf(format, a...))
}

func LogD(format string, a ...interface{}) {
	fmt.Printf("[%s][debug] %s\n", time.Now().Local().Format(time.RFC3339), fmt.Sprintf(format, a...))
}

func LogW(format string, a ...interface{}) {
	fmt.Printf("[%s][warn] %s\n", time.Now().Local().Format(time.RFC3339), fmt.Sprintf(format, a...))
}

func LogE(format string, a ...interface{}) {
	fmt.Printf("[%s][error] %s\n", time.Now().Local().Format(time.RFC3339), fmt.Sprintf(format, a...))
}
