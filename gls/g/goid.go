package g

import (
	"unsafe"
)

func getg() uintptr

// G returns current g (the goroutine struct) to user space.
//
//go:nosplit
func G() unsafe.Pointer {
	return unsafe.Pointer(getg())
}
