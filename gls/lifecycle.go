package gls

import (
	"fmt"
	"github.com/funbytes/modern-go/gls/g"
	"reflect"
	"runtime"
	"sync/atomic"
	"unsafe"
)

// GStatus represents the real status of runtime.g
type GStatus uint32

const (
	GIdle      GStatus = 0 // see runtime._Gidle
	GRunnable  GStatus = 1 // see runtime._Grunnable
	GRunning   GStatus = 2 // see runtime._Grunning
	GSyscall   GStatus = 3 // see runtime._Gsyscall
	GWaiting   GStatus = 4 // see runtime._Gwaiting
	GMoribund  GStatus = 5 // see runtime._Gmoribund_unused
	GDead      GStatus = 6 // see runtime._Gdead
	GEnqueue   GStatus = 7 // see runtime._Genqueue_unused
	GCopystack GStatus = 8 // see runtime._Gcopystack
	GPreempted GStatus = 9 // see runtime._Gpreempted
)

var (
	goidOffset   uintptr
	labelsOffset uintptr
	statusOffset uintptr
)

func init() {
	offset := func(t reflect.Type, f string) uintptr {
		if field, found := t.FieldByName(f); found {
			return field.Offset
		}
		panic(fmt.Sprintf("init routine failed, cannot find g.%s, version=%s", f, runtime.Version()))
	}
	gt := reflect.TypeOf(g.G0())
	goidOffset = offset(gt, "goid")
	labelsOffset = offset(gt, "labels")
	statusOffset = offset(gt, "atomicstatus")
}

func routineLabels(gp unsafe.Pointer) map[string]string {
	labelsPPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(gp) + labelsOffset))
	labelsPtr := atomic.LoadPointer(labelsPPtr)
	if labelsPtr == nil {
		return nil
	}
	// see SetGoroutineLabels, labelsPtr is `*labelMap`
	return *(*map[string]string)(labelsPtr)
}

func routineStatus(gp unsafe.Pointer) GStatus {
	statusPtr := (*uint32)(unsafe.Pointer(uintptr(gp) + statusOffset))
	return GStatus(atomic.LoadUint32(statusPtr))
}

func routineSetLabels(gp unsafe.Pointer, labels map[string]string) {
	labelsPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(gp) + labelsOffset))
	atomic.StorePointer(labelsPtr, unsafe.Pointer(&labels))
}

func routineHack(se *slotElem, gp unsafe.Pointer) bool {
	registerFinalizer(se, gp)
	return true
}

func routineUnhack(gp unsafe.Pointer) {
}

// register Register finalizer into goroutine's lifecycle
func registerFinalizer(se *slotElem, gp unsafe.Pointer) {
	labels := make(map[string]string)
	for k, v := range routineLabels(gp) {
		labels[k] = v
	}
	runtime.SetFinalizer(&labels, func(_ interface{}) {
		finalize(se, gp)
	})
	routineSetLabels(gp, labels)
}

var Cnt int

func finalize(se *slotElem, gp unsafe.Pointer) {
	if gp == nil || se == nil {
		return
	}

	// Maybe others (pprof) replaced our labels, register it again.
	if routineStatus(gp) != GDead {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					errLog(fmt.Sprintf("store.finalize panic error: %v", err))
				}
			}()

			registerFinalizer(se, gp)
			Cnt++
		}()

		return
	}

	resetAtExit(se, gp)
}
