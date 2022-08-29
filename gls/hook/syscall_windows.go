// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

//go:build !go1.17
// +build !go1.17

package hook

import (
	"fmt"
	"golang.org/x/sys/windows"
	"unsafe"
)

const (
	protectRead  = windows.PAGE_READONLY
	protectWrite = windows.PAGE_READWRITE
)

func mprotect(ptr unsafe.Pointer, size, prot uintptr) {
	var oldprotect uint32
	err := windows.VirtualProtect(uintptr(ptr), size, uint32(prot), &oldprotect)
	if err != nil {
		panic(fmt.Errorf("tls: fail to call VirtualProtect(addr=0x%x, size=%v, prot=0x%x) with error %v", ptr, size, prot, err))
	}
}
