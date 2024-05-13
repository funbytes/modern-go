// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

// Package g exposes goroutine struct g to user space.
package g

import (
	"reflect"
	"unsafe"
)

// g0 the value of runtime.g0.
//
//go:linkname g0 runtime.g0
var g0 struct{}

// getgp returns the pointer to the current runtime.g.
//
//go:nosplit
func getgp() unsafe.Pointer

// getg0 returns the value of runtime.g0.
//
//go:nosplit
func getg0() interface{} {
	return packEface(getgt(), unsafe.Pointer(&g0))
}

// getgt returns the type of runtime.g.
//
//go:nosplit
func getgt() reflect.Type {
	return typeByString("runtime.g")
}

// G returns current g (the goroutine struct) to user space.
//
//go:nosplit
func G() unsafe.Pointer {
	return getgp()
}

// G0 returns the g0, the main goroutine.
//
//go:nosplit
func G0() interface{} {
	return getg0()
}
