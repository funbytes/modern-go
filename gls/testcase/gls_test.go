// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package testcase

import (
	"fmt"
	"github.com/funbytes/modern-go/gls"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

type tlsKey1 struct{}
type tlsKey2 struct{}
type tlsKey3 struct{}

type payload struct {
	data [1024]byte
}

func triggerMoreStack(n int, p payload) int {
	if n <= 0 {
		return 0
	}

	// Avoid tail optimization.
	p.data[n] = 1
	return triggerMoreStack(n-1, p) + int(p.data[0]+p.data[len(p.data)-1])
}

type closerFunc func()

func (f closerFunc) Close() error {
	f()
	return nil
}

func TestTLS(t *testing.T) {
	times := 1000
	idMap := map[unsafe.Pointer]struct{}{}
	idMapMu := sync.Mutex{}

	for i := 0; i < times; i++ {
		t.Run(fmt.Sprintf("Round %v", i), func(t *testing.T) {
			t.Parallel()
			closed := false
			k1 := tlsKey1{}
			v1 := 1234
			k2 := tlsKey2{}
			v2 := "v2"
			k3 := tlsKey3{}
			v3 := closerFunc(func() {
				closed = true
			})
			cnt := 0

			gls.Set(k1, gls.MakeData(v1))
			gls.Set(k2, gls.MakeData(v2))
			gls.Set(k3, gls.MakeData(v3))

			cnt++
			gls.AtExit(func() {
				cnt--

				if expected := 0; cnt != expected {
					t.Fatalf("AtExit should call func in FILO order.")
				}
			})

			cnt++
			gls.AtExit(func() {
				cnt--

				if expected := 1; cnt != expected {
					t.Fatalf("AtExit should call func in FILO order.")
				}
			})

			if d, ok := gls.Get(k1); !ok || d == nil || !reflect.DeepEqual(d.Value(), v1) {
				t.Fatalf("fail to get k1.")
			}

			if d, ok := gls.Get(k2); !ok || d == nil || !reflect.DeepEqual(d.Value(), v2) {
				t.Fatalf("fail to get k2.")
			}

			triggerMoreStack(1000, payload{})

			gls.Reset()

			if !closed {
				t.Fatalf("v3.Close() is not called.")
			}

			if _, ok := gls.Get(k1); ok {
				t.Fatalf("k1 should be empty.")
			}

			gls.Set(k1, gls.MakeData(v1))
			gls.Set(k1, gls.MakeData(v2))

			if d, ok := gls.Get(k1); !ok || d == nil || !reflect.DeepEqual(d.Value(), v2) {
				t.Fatalf("fail to get k1.")
			}

			if _, ok := gls.Get(k2); ok {
				t.Fatalf("k2 should be empty.")
			}

			cnt++
			gls.AtExit(func() {
				cnt--

				if expected := 2; cnt != expected {
					t.Fatalf("AtExit should call func in FILO order.")
				}
			})

			id := gls.ID()

			if id == nil {
				t.Fatalf("fail to get ID. [id:%v]", id)
			}

			idMapMu.Lock()
			defer idMapMu.Unlock()

			if _, ok := idMap[id]; ok {
				t.Fatalf("duplicated ID. [id:%v]", id)
			}

			idMap[id] = struct{}{}
		})
	}
}

func TestUnload(t *testing.T) {
	// Run test in a standalone goroutine.
	t.Run("try unload", func(t *testing.T) {
		t.Parallel()
		id := gls.ID()
		exitCalled := false
		gls.AtExit(func() {
			exitCalled = true
		})
		key := "key"
		expected := "value"
		gls.Set(key, gls.MakeData(expected))

		if d, ok := gls.Get(key); !ok {
			t.Fatalf("fail to get data. [key:%v]", key)
		} else if actual, ok := d.Value().(string); !ok || actual != expected {
			t.Fatalf("invalid value. [key:%v] [value:%v] [expected:%v]", key, actual, expected)
		}

		gls.Unload()

		// It's ok to call it again.
		gls.Unload()

		if gls.IsGlsEnabled(gls.ID()) {
			t.Fatalf("id must be changed after unload. [id:%v]", id)
		}

		if _, ok := gls.Get(key); ok {
			t.Fatalf("key must be cleared. [key:%v]", key)
		}

		if exitCalled {
			t.Fatalf("all AtExit functions must not be called.")
		}
	})
}
func TestShrinkStack(t *testing.T) {
	const times = 500
	const gcTimes = 100
	sleep := 100 * time.Microsecond
	errors := make(chan error, times)
	var done int64

	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup
	wg.Add(times)

	for i := 0; i < times; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					errors <- fmt.Errorf("recovered with message: %v", r)
				}
			}()

			gls.AtExit(func() {
				atomic.AddInt64(&done, 1)
				wg.Done()
			})
			n := rand.Intn(gcTimes)

			for j := 0; j < n; j++ {
				triggerMoreStack(100, payload{})
				time.Sleep(time.Duration((0.5 + rand.Float64()) * float64(sleep)))
			}
		}()
	}

	exit := make(chan bool, 2)
	go func() {
		wg.Wait()
		exit <- true
	}()

	go func() {
		// Avoid deadloop.
		select {
		case <-time.After(15 * time.Second):
			exit <- false
		}
	}()

GC:
	for {
		time.Sleep(sleep)
		runtime.GC()

		select {
		case <-exit:
			break GC
		default:
		}
	}

	failed := false

DumpError:
	for {
		select {
		case err := <-errors:
			failed = true
			t.Logf("panic [err:%v]", err)
		default:
			break DumpError
		}
	}

	if failed {
		t.FailNow()
	}

	runtime.GC()
	time.Sleep(5 * time.Minute)

	t.Logf("kkkkkkk finalize count:%d", gls.Cnt)
	if done != times {
		t.Fatalf("some AtExit handlers are not called. [expected:%v] [actual:%v]", times, done)
	}
}

func TestUnloadInAtExitHandker(t *testing.T) {
	ch := make(chan bool, 1)
	go func() {
		gls.AtExit(func() {
			gls.Unload()
		})
		ch <- true
	}()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic. [r:%v]", r)
		}
	}()
	<-ch
}
