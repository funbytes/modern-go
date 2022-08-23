// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package hook

import (
	"fmt"
	"github.com/funbytes/modern-go/gls/c"
	"syscall"
	"unsafe"
)

const (
	protectRead  = syscall.PROT_READ
	protectWrite = syscall.PROT_READ | syscall.PROT_WRITE
)

func mprotect(ptr unsafe.Pointer, size, prot uintptr) {
	err := c.Unprotect(ptr, size, prot)
	if err != nil {
		panic(fmt.Errorf("tls: fail to call mprotect(addr=0x%x, size=%v, prot=0x%x) with error:%v", uintptr(ptr), size, prot, err))
	}
}
