//go:build !go1.17
// +build !go1.17

package gls

import (
	"github.com/funbytes/modern-go/gls/g"
	"github.com/funbytes/modern-go/gls/hook"
	"unsafe"
)

func init() {
	hook.ResetAtExit = exitWithReset
}

func routineHack(se *slotElem, gp unsafe.Pointer) bool {
	return hook.Hack(gp)
}

func routineUnhack(gp unsafe.Pointer) {
	hook.Unhack(gp)
}

func exitWithReset() {
	gp := g.G()
	se := findSlot(gp)
	resetAtExit(se, gp)
}
