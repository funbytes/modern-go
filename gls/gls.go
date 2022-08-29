package gls

import (
	"github.com/funbytes/modern-go/gls/g"
	"unsafe"
)

func SetErrorLog(l func(string)) {
	if l != nil {
		errLog = l
	}
}

// Get data by key.
func Get(key interface{}) (d Data, ok bool) {
	dm := fetchDataMap(true)

	if dm == nil {
		return
	}

	d, ok = dm.data[key]
	return
}

// Set data for key.
func Set(key interface{}, data Data) {
	dm := fetchDataMap(false)
	dm.data[key] = data
}

// Del data by key.
func Del(key interface{}) {
	dm := fetchDataMap(true)

	if dm == nil {
		return
	}

	delete(dm.data, key)
}

// ID returns a unique ID for a goroutine.
// If it's not possible to get the value, ID returns 0.
//
// It's guaranteed to be unique and consistent for one goroutine,
// unless it's called after Unload, which completely resets TLS stub.
// To be clear, it's not goid used by Go runtime.
func ID() unsafe.Pointer {
	return g.G()
}

// IsGlsEnabled test if the gls is available for specified goroutine
func IsGlsEnabled(id unsafe.Pointer) bool {
	if id == InvalidID {
		return false
	}

	se := findSlot(id)
	if se == nil {
		return false
	}

	_, ok := getGlsData(se, id)
	return ok
}

// AtExit runs f when current goroutine is exiting.
// The f is called in FILO order.
func AtExit(f func()) {
	dm := fetchDataMap(false)
	dm.atExitFuncs = append(dm.atExitFuncs, f)
}

// Reset clears TLS data and releases all resources for current goroutine.
// It doesn't remove any AtExit handlers.
func Reset() {
	se, gp := getSlotElem()
	if se == nil || gp == nil {
		return
	}

	reset(se, gp, false)
}

// Unload completely unloads TLS and clear all data and AtExit handlers.
func Unload() {
	se, gp := getSlotElem()
	if se == nil || gp == nil {
		return
	}

	if !reset(se, gp, true) {
		routineUnhack(gp)
	}
}
