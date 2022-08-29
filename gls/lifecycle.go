package gls

import (
	"fmt"
	"github.com/funbytes/modern-go/gls/g"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
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

// labelMap is the representation of the label set held in the context type.
// This is an initial implementation, but it will be replaced with something
// that admits incremental immutable modification more efficiently.
type labelMap map[string]string

var (
	goidOffset   uintptr
	labelsOffset uintptr
	statusOffset uintptr
)

const finalizedLabel = "modern-go-gls-support-go1.17-or-higher"

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

// Goid get the unique goid of the current routine.
func routineGoId(gp unsafe.Pointer) int64 {
	goidPtr := (*int64)(unsafe.Pointer(uintptr(gp) + goidOffset))
	return atomic.LoadInt64(goidPtr)
}

func routineLabels(gp unsafe.Pointer) labelMap {
	labelsPPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(gp) + labelsOffset))
	labelsPtr := atomic.LoadPointer(labelsPPtr)
	if labelsPtr == nil {
		return nil
	}
	// see SetGoroutineLabels, labelsPtr is `*labelMap`
	return *(*labelMap)(labelsPtr)
}

func routineStatus(gp unsafe.Pointer) GStatus {
	statusPtr := (*uint32)(unsafe.Pointer(uintptr(gp) + statusOffset))
	return GStatus(atomic.LoadUint32(statusPtr))
}

func routineSetLabels(gp unsafe.Pointer, labels labelMap) {
	if _, ok := labels[finalizedLabel]; !ok {
		labels[finalizedLabel] = "hello world"
	}
	labelsPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(gp) + labelsOffset))
	atomic.StorePointer(labelsPtr, unsafe.Pointer(&labels))
}

func routineHack(se *slotElem, gp unsafe.Pointer) bool {
	registerFinalizer(se, gp)
	return true
}

func routineUnhack(gp unsafe.Pointer) {
}

// register Register finalizer into goroutine's lifeCycle
func registerFinalizer(se *slotElem, gp unsafe.Pointer) bool {
	if routineStatus(gp) == GDead {
		return false
	}

	labels := routineLabels(gp)
	_, ok := labels[finalizedLabel]
	if labels == nil || !ok {
		labels = make(labelMap)
		routineSetLabels(gp, labels)
		runtime.SetFinalizer(&labels, func(_ interface{}) {
			finalize(se, gp)
		})
	}

	return routineStatus(gp) != GDead
}

var Cnt atomic.Int32
var DeadCnt atomic.Int32

func finalize(se *slotElem, gp unsafe.Pointer) {
	if gp == nil || se == nil {
		return
	}

	// Maybe others (pprof) replaced our labels, register it again.
	status := routineStatus(gp)
	if status != GDead {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					errLog(fmt.Sprintf("store.finalize panic error: %v", err))
				}
			}()

			Cnt.Add(1)
			time.Sleep(10 * time.Millisecond)
			if !registerFinalizer(se, gp) {
				DeadCnt.Add(1)
				resetAtExit(se, gp)
			}
		}()
		return
	}

	DeadCnt.Add(1)
	resetAtExit(se, gp)
}
