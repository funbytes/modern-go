package tls

import (
	"runtime"
	"time"
	"unsafe"

	"github.com/modern-go/reflect2"
	"gitlab-ee.funplus.io/watcher/watcher/misc/gotls/g"
	"gitlab-ee.funplus.io/watcher/watcher/misc/wpool"
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
	wpool.Go(func() {
		resetAtExit(id)
	})
}
