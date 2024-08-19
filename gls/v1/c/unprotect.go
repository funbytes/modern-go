package c

/*
#include <stddef.h>
#include <errno.h>
#include <stdio.h>
#include <unistd.h>
#include <sys/mman.h>
#include <mach/mach.h>

int subhook_unprotect(void *address, long pagesize, size_t size, size_t flag) {
	address = (void *)((long)address & ~(pagesize - 1));
	int error = mprotect(address, size, flag);
	if (-1 == error) {
        // If mprotect fails, try to use VM_PROT_COPY with vm_protect
		kern_return_t kret = vm_protect(mach_task_self(), (unsigned long)address, size, 0, flag | VM_PROT_COPY);
		if (kret != KERN_SUCCESS) {
			fprintf(stderr, "vm_protect error! kern_return_t:%d errno:%d\n", kret, errno);
			return -1;
		}
		return 0;
	}
	return error;
}
*/
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"
)

func Unprotect(ptr unsafe.Pointer, size, prot uintptr) error {
	errno := C.subhook_unprotect(ptr, C.long(syscall.Getpagesize()), C.size_t(size), C.size_t(prot))
	if int(errno) != 0 {
		return fmt.Errorf("unprotect hook error:%d", int(errno))
	}
	return nil
}
