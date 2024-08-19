// Package tls creates a TLS for a goroutine and release all resources at goroutine exit.
package tls

import (
	"io"

	"gitlab-ee.funplus.io/watcher/watcher/misc/concurrent"
	"gitlab-ee.funplus.io/watcher/watcher/misc/gotls/g"
)

const goroutineCount = 10240

var (
	tlsDataMap = concurrent.StaticBucketAnyValueHashMap[int64, *tlsData]{}
)

type tlsData struct {
	data        dataMap
	kv          map[any]any
	glsBase     any
	atExitFuncs []func()
}

type dataMap map[any]Data

func init() {
	tlsDataMap.Init(goroutineCount)
}

// Get data by key.
func Get(key any) (d Data, ok bool) {
	id := ID()
	dm, _ := tlsDataMap.Get(id)
	if dm != nil {
		d, ok = dm.data[key]
		return
	}
	return nil, false
}

// Set data for key.
func Set(key any, data Data) {
	dm := fetchDataMap()
	dm.data[key] = data
}

func SetGLS(data any) {
	dm := fetchDataMap()
	dm.glsBase = data
}

func GetGLS[T any]() *T {
	id := ID()
	dm, _ := tlsDataMap.Get(id)
	if dm == nil || dm.glsBase == nil {
		return nil
	}
	v, _ := dm.glsBase.(*T)
	return v
}

func SetKV(key any, data any) {
	dm := fetchDataMap()
	dm.kv[key] = data
}

func GetKV[T any](key any) (ret T, ok bool) {
	id := ID()
	dm, _ := tlsDataMap.Get(id)
	if dm != nil {
		d, ok := dm.kv[key]
		if ok {
			ret, ok = d.(T)
			return ret, ok
		}
	}
	return ret, false
}

func DelKV(key any) {
	id := ID()
	dm, _ := tlsDataMap.Get(id)
	if dm == nil {
		return
	}
	delete(dm.kv, key)
}

// Del data by key.
func Del(key any) {
	id := ID()
	dm, _ := tlsDataMap.Get(id)
	if dm == nil {
		return
	}
	delete(dm.data, key)
}

// ID returns a unique ID for a goroutine.
//
//go:nosplit
func ID() int64 {
	return routineID(g.G())
}

// AtExit runs f when current goroutine is exiting.
// The f is called in FILO order.
func AtExit(f func()) {
	dm := fetchDataMap()
	dm.atExitFuncs = append(dm.atExitFuncs, f)
}

// AtExitID runs f when current goroutine is exiting.
// The f is called in FILO order.
func AtExitID(f func(goId int64)) {
	dm := fetchDataMap()
	id := ID()
	dm.atExitFuncs = append(dm.atExitFuncs, func() {
		f(id)
	})
}

// Reset clears TLS data and releases all resources for current goroutine.
// It doesn't remove any AtExit handlers.
func Reset() {
	reset(false)
}

func reset(complete bool) {
	var data dataMap
	var needUnHack bool
	id := ID()

	if complete {
		dm, ok := tlsDataMap.Del(id)
		if ok {
			data = dm.data
			needUnHack = true
		}
	} else {
		dm, ok := tlsDataMap.Get(id)
		if ok {
			data = dm.data
			dm.data = dataMap{}
			dm.kv = map[any]any{}
			dm.glsBase = nil
		}
	}

	if needUnHack {
		unregisterFinalizer()
	}

	for _, d := range data {
		safeClose(d)
	}
}

// Unload completely unloads TLS and clear all data and AtExit handlers.
func Unload() {
	reset(true)
}

func resetAtExit(id int64) {
	dm, ok := tlsDataMap.Del(id)
	if !ok {
		return
	}

	funcs := dm.atExitFuncs
	// Call handlers in FILO order.
	for i := len(funcs) - 1; i >= 0; i-- {
		safeRun(funcs[i])
	}

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

func fetchDataMap() *tlsData {
	id := ID()

	// Try to find saved data.
	dm, get := tlsDataMap.GetOrInsert(id, func(int64) *tlsData {
		return &tlsData{
			data: dataMap{},
			kv:   map[any]any{},
		}
	})
	// Current goroutine is not hacked. Hack it.
	if !get {
		registerFinalizer(id, g.G())
	}
	return dm
}
