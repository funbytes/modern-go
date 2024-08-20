package utils

import (
	"context"
	"fmt"
	"runtime/pprof"
	"strings"

	"github.com/go-stack/stack"
)

func GetFileAndFuncName() string {
	return getFileAndFuncName(5)
}

func getFileAndFuncName(skip int) string {
	c := stack.Caller(skip)
	filename := c.Frame().File
	filename = filename[strings.LastIndex(filename, "/")+1:]
	funcname := c.Frame().Function
	var globalGeneric bool
	if strings.HasSuffix(funcname, "[...]") {
		globalGeneric = true
		funcname = funcname[:len(funcname)-5]
	}
	index := strings.LastIndex(funcname, ".")
	if index != -1 {
		funcname = funcname[index+1:]
	}
	line := c.Frame().Line
	if globalGeneric {
		return fmt.Sprintf("%s:%d:%s[]", filename, line, funcname)
	} else {
		return fmt.Sprintf("%s:%d:%s", filename, line, funcname)
	}
}

// Labels generates the labels for a function/method
type Labels func() []string

// SetLabels will set debugger labels for any function/method call
func SetLabels(l Labels) {
	pprof.SetGoroutineLabels(pprof.WithLabels(context.Background(), pprof.Labels(l()...)))
}
