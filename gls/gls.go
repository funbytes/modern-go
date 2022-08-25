// Copyright 2020 yc. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

//go:build go1.7
// +build go1.7

// Package tls creates a GLS for a goroutine and release all resources at goroutine exit.
package gls

import (
	"io"
	"sync"
	"unsafe"

	"github.com/funbytes/modern-go/gls/g"
)

const shardsCount = 31

var (
	InvalidID   unsafe.Pointer = nil
	globalSlots []*slotElem
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
	globalSlots = make([]*slotElem, shardsCount)
	for i := 0; i < shardsCount; i++ {
		globalSlots[i] = &slotElem{
			dataMap: make(glsMapType),
		}
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

// Unload completely unloads TLS and clear all data and AtExit handlers.
func Unload() {
	se, gp := getSlotElem()
	if se == nil || gp == nil {
		return
	}

	if !reset(se, gp, true) {
		unhack(gp)
	}
}

func resetAtExit() {
	se, gp := getSlotElem()
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
		if !hack(gp) {
			delGlsData(se, gp)
		}
	}

	return dm
}

func findSlot(gp unsafe.Pointer) *slotElem {
	if gp == InvalidID {
		return nil
	}

	gpid := uintptr(gp)
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
