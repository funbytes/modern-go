package gls

import (
	"fmt"
	"github.com/funbytes/modern-go/gls/g"
	"reflect"
	"runtime"
	"sync"
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

var (
	lifeCycle *gcGuard
)

const goLabelsKey = "modern-go-gls-support-go1.17-or-higher"
const lifeCycleGCInterval = time.Second * 1 // The pre-defined gc interval

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

	// // life cycle
	// lifeCycle = &gcGuard{
	// 	lcGCTimer: time.AfterFunc(lifeCycleGCInterval, clearDeadStore),
	// }
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
	labelsPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(gp) + labelsOffset))
	atomic.StorePointer(labelsPtr, unsafe.Pointer(&labels))
}

func routineHack(se *slotElem, gp unsafe.Pointer) bool {
	registerFinalizer(se, gp)
	return true
}

func routineUnhack(gp unsafe.Pointer) {
}

type gcGuard struct {
	lcLock    sync.Mutex  // The Lock to control accessing of storages
	lcGCTimer *time.Timer // The timer of storage's garbage collector
	lcIndex   uint        // Index for globalSlots
}

// func registerFinalizer(se *slotElem, gp unsafe.Pointer) {
// 	// lifeCycle.lcLock.Lock()
// 	// lifeCycle.lcLock.Unlock()
// 	// if lifeCycle.lcGCTimer == nil {
// 	// 	lifeCycle.lcGCTimer = time.AfterFunc(lifeCycleGCInterval, clearDeadStore)
// 	// }
// }

// type slotPair struct {
// 	se *slotElem
// 	gp unsafe.Pointer
// }

// var Cnt atomic.Int32
// var DeadCnt atomic.Int32

// func clearDeadStore() {
// 	Cnt.Add(1)
//
// 	var allCnt, deadCnt int
// 	var gplist []*slotPair
//
// 	// lifeCycle.lcLock.Lock()
// 	// lifeCycle.lcIndex++
// 	// index := lifeCycle.lcIndex
// 	// lifeCycle.lcLock.Unlock()
//
// 	for i := 0; i < shardsCount; i++ {
// 		se := globalSlots[i]
// 		se.rwlock.RLock()
// 		for k, _ := range se.dataMap {
// 			allCnt++
// 			gplist = append(gplist, &slotPair{
// 				se: se,
// 				gp: k,
// 			})
// 		}
// 		se.rwlock.RUnlock()
// 	}
//
// 	var stat []GStatus
// 	for _, v := range gplist {
// 		status := routineStatus(v.gp)
// 		if status == GDead {
// 			deadCnt++
// 			DeadCnt.Add(1)
// 			resetAtExit(v.se, v.gp)
// 		} else {
// 			stat = append(stat, status)
// 		}
// 	}
//
// 	fmt.Printf("allCnt:%d deadCnt:%d status:%v \n", allCnt, deadCnt, stat)
//
// 	lifeCycle.lcLock.Lock()
// 	defer lifeCycle.lcLock.Unlock()
// 	lifeCycle.lcGCTimer.Reset(lifeCycleGCInterval)
// 	// if allCnt > deadCnt {
// 	// 	lifeCycle.lcGCTimer.Reset(lifeCycleGCInterval)
// 	// } else {
// 	// 	lifeCycle.lcGCTimer = nil
// 	// }
// }

// register Register finalizer into goroutine's lifeCycle
func registerFinalizer(se *slotElem, gp unsafe.Pointer) {
	if routineStatus(gp) == GDead {
		return
	}

	labels := routineLabels(gp)
	if labels == nil {
		labels = make(labelMap)
		routineSetLabels(gp, labels)
		se.rwlock.Lock()
		defer se.rwlock.Unlock()
		se.labels = uintptr(unsafe.Pointer(&labels))
		runtime.SetFinalizer(&labels, func(_ interface{}) {
			finalize(se, gp)
		})
	} else if _, ok := getGlsData(se, gp); ok {
		se.rwlock.Lock()
		defer se.rwlock.Unlock()
		if se.labels != uintptr(unsafe.Pointer(&labels)) {
			runtime.SetFinalizer(&labels, func(_ interface{}) {
				finalize(se, gp)
			})
		}
	}
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

			registerFinalizer(se, gp)
			Cnt.Add(1)
		}()

		return
	}

	DeadCnt.Add(1)
	resetAtExit(se, gp)
}
