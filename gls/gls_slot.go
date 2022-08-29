// Copyright 2020 yc. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

//go:build go1.7
// +build go1.7

// Package tls creates a GLS for a goroutine and release all resources at goroutine exit.
package gls

import (
	"fmt"
	"io"
	"os"
	"sync"
	"unsafe"

	"github.com/funbytes/modern-go/gls/g"
)

const shardsCount = 31

var (
	InvalidID   unsafe.Pointer = nil
	globalSlots []*slotElem
	once        sync.Once
	errLog      func(string)
)

type glsMapType map[unsafe.Pointer]*glsData
type dataMapType map[interface{}]Data

type slotElem struct {
	rwlock  sync.RWMutex
	dataMap glsMapType
}

type glsData struct {
	data        dataMapType
	atExitFuncs []func()
	done        bool
}

// As we cannot hack main goroutine safely,
// proactively create TLS for main to avoid hacking.
func init() {
	gp := g.G()
	if gp == nil {
		return
	}

	once.Do(func() {
		globalSlots = make([]*slotElem, shardsCount)
		for i := 0; i < shardsCount; i++ {
			globalSlots[i] = &slotElem{
				dataMap: make(glsMapType),
			}
		}
		errLog = func(s string) {
			_, _ = fmt.Fprintf(os.Stderr, s)
		}
	})
}

func reset(se *slotElem, gp unsafe.Pointer, complete bool) (alreadyReset bool) {
	var data dataMapType
	dm, ok := getGlsData(se, gp)
	if !ok {
		return
	}

	if dm == nil || dm.done {
		alreadyReset = true
	} else {
		data = dm.data
		if complete {
			delGlsData(se, gp)
		} else {
			dm.data = make(dataMapType)
		}
	}

	for _, d := range data {
		safeClose(d)
	}
	return
}

func resetAtExit(se *slotElem, gp unsafe.Pointer) {
	if se == nil || gp == nil {
		return
	}

	dm, ok := getGlsData(se, gp)
	if !ok {
		return
	}

	funcs := dm.atExitFuncs
	dm.atExitFuncs = nil
	dm.done = true

	// Call handlers in FILO order.
	for i := len(funcs) - 1; i >= 0; i-- {
		safeRun(funcs[i])
	}

	delGlsData(se, gp)

	for _, d := range dm.data {
		safeClose(d)
	}
}

// safeRun runs f and ignores any panic.
func safeRun(f func()) {
	defer func() {
		recover()
	}()
	f()
}

// safeClose closes closer and ignores any panic.
func safeClose(closer io.Closer) {
	defer func() {
		recover()
	}()
	closer.Close()
}

func fetchDataMap(readonly bool) *glsData {
	se, gp := getSlotElem()
	if se == nil || gp == nil {
		return nil
	}

	// Try to find saved data.
	needHack := false
	dm, _ := getGlsData(se, gp)
	if dm == nil && !readonly {
		needHack = true
		dm = newGlsData(se, gp)
	}

	// Current goroutine is not hacked. Hack it.
	if needHack {
		if !routineHack(se, gp) {
			delGlsData(se, gp)
		}
	}

	return dm
}

func findSlot(gp unsafe.Pointer) *slotElem {
	if gp == InvalidID {
		return nil
	}

	gpid := routineGoId(gp)
	shardIndex := gpid % shardsCount
	return globalSlots[shardIndex]
}

func getSlotElem() (*slotElem, unsafe.Pointer) {
	gp := g.G()
	if gp == nil {
		return nil, nil
	}

	se := findSlot(gp)
	if se == nil {
		return nil, nil
	}
	return se, gp
}

func getGlsData(se *slotElem, gp unsafe.Pointer) (*glsData, bool) {
	se.rwlock.RLock()
	defer se.rwlock.RUnlock()
	dm, ok := se.dataMap[gp]
	return dm, ok
}

func newGlsData(se *slotElem, gp unsafe.Pointer) *glsData {
	dm := &glsData{
		data: make(dataMapType),
	}

	se.rwlock.Lock()
	defer se.rwlock.Unlock()
	se.dataMap[gp] = dm
	return dm
}

func delGlsData(se *slotElem, gp unsafe.Pointer) {
	se.rwlock.Lock()
	defer se.rwlock.Unlock()
	delete(se.dataMap, gp)
}
