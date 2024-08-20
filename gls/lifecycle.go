package gls

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"github.com/funbytes/modern-go/gls/g"
	"github.com/modern-go/reflect2"
)

var (
	goidOffset  uintptr
	timerOffset uintptr
)

func init() {
	gType := reflect2.TypeByName("runtime.g").(reflect2.StructType)
	if gType == nil {
		panic("failed to get runtime.g type")
	}
	goidOffset = gType.FieldByName("goid").Offset()
	timerOffset = gType.FieldByName("timer").Offset()
}

func routineTimer(gp unsafe.Pointer) *struct{} {
	return (*struct{})(*(*unsafe.Pointer)(unsafe.Add(gp, timerOffset)))
}

//go:nosplit
func routineID(gp unsafe.Pointer) int64 {
	return *(*int64)(unsafe.Add(gp, goidOffset))
}

// register Register finalizer into goroutine's lifeCycle
func registerFinalizer(id int64, gp unsafe.Pointer) bool {
	gTimer := routineTimer(gp)
	if gTimer == nil {
		time.Sleep(time.Nanosecond)
		gTimer = routineTimer(gp)
		if gTimer == nil {
			panic("cant found the timer")
		}
	}
	runtime.SetFinalizer(gTimer, func(_ any) {
		finalize(id)
	})
	return true
}

func unregisterFinalizer() {
	gTimer := routineTimer(g.G())
	if gTimer != nil {
		runtime.SetFinalizer(gTimer, nil)
	}
}

func finalize(id int64) {
	// Maybe others (pprof) replaced our labels, register it again.
	go func() {
		defer func() {
			if err := recover(); err != nil {
				errLog(fmt.Sprintf("store.finalize panic error: %v", err))
			}
		}()
		resetAtExit(id)
	}()
}
