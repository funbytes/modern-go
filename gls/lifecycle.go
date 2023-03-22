//go:build go1.17
// +build go1.17

package gls

import (
	"fmt"
	"runtime"
	"strconv"
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

// labelMap is the representation of the label set held in the context type.
// This is an initial implementation, but it will be replaced with something
// that admits incremental immutable modification more efficiently.
type labelMap map[string]string

const goLabelsKey = "[=-*(goLabelsKey)*-=]"

func routineLabelPtr(gp unsafe.Pointer) (labelMap, unsafe.Pointer) {
	labelsPPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(gp) + labelsOffset))
	labelsPtr := atomic.LoadPointer(labelsPPtr)
	if labelsPtr == nil {
		return nil, labelsPtr
	}
	// see SetGoroutineLabels, labelsPtr is `*labelMap`
	return *(*labelMap)(labelsPtr), labelsPtr
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

func routineID(gp unsafe.Pointer) int64 {
	return *(*int64)(unsafe.Pointer(uintptr(gp) + goidOffset))
}

func routineStatus(gp unsafe.Pointer) GStatus {
	statusPtr := (*uint32)(unsafe.Pointer(uintptr(gp) + statusOffset))
	return GStatus(atomic.LoadUint32(statusPtr))
}

func routineSetLabels(gp unsafe.Pointer, old unsafe.Pointer, labels *labelMap) bool {
	labelsPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(gp) + labelsOffset))
	return atomic.CompareAndSwapPointer(labelsPtr, old, unsafe.Pointer(labels))
}

func routineHack(se *slotElem, gp unsafe.Pointer) bool {
	return registerFinalizer(se, gp, true)
}

func routineUnhack(gp unsafe.Pointer) {
}

// register Register finalizer into goroutine's lifeCycle
func registerFinalizer(se *slotElem, gp unsafe.Pointer, inGoroutine bool) bool {
	id := routineGoId(gp)
	labels, old := routineLabelPtr(gp)
	if !inGoroutine {
		if routineStatus(gp) == GDead || routineID(gp) != id || labels == nil {
			return false
		}
	}
	idStr := strconv.Itoa(int(id))
	oldID, ok := labels[goLabelsKey]
	if ok {
		// Has already been registered and does not need to be processed again
		if oldID == idStr {
			return true
		} else if !inGoroutine {
			return false
		}
	}
	// Copy the label, re-register, and ensure that the label is read-only
	nLabels := make(labelMap)
	for k, v := range labels {
		nLabels[k] = v
	}
	nLabels[goLabelsKey] = idStr
	nPtr := &nLabels
	if !routineSetLabels(gp, old, nPtr) {
		// 说明是在goroutine外执行，有修改，递归执行一次
		return registerFinalizer(se, gp, inGoroutine)
	}
	runtime.SetFinalizer(nPtr, func(_ interface{}) {
		finalize(se, gp)
	})
	return true
}

func finalize(se *slotElem, gp unsafe.Pointer) {
	id := routineGoId(gp)
	if routineStatus(gp) == GDead || routineID(gp) != id || routineLabels(gp) == nil {
		resetAtExit(se, gp)
		return
	}
	// Maybe others (pprof) replaced our labels, register it again.
	go func() {
		defer func() {
			if err := recover(); err != nil {
				errLog(fmt.Sprintf("store.finalize panic error: %v", err))
			}
		}()

		if !registerFinalizer(se, gp, false) {
			resetAtExit(se, gp)
		}
	}()
}
