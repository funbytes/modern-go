package gls

import (
	"fmt"
	"github.com/funbytes/modern-go/gls/g"
	"reflect"
	"runtime"
	"sync/atomic"
	"unsafe"
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

// Goid get the unique goid of the current routine.
func routineGoId(gp unsafe.Pointer) int64 {
	goidPtr := (*int64)(unsafe.Pointer(uintptr(gp) + goidOffset))
	return atomic.LoadInt64(goidPtr)
}
