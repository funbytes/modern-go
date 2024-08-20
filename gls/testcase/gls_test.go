package testcase

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/funbytes/modern-go/gls"
	"github.com/funbytes/modern-go/gls/g"
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
	return triggerMoreStack(n-1, p) + int(p.data[0]+p.data[len(p.data)-1])
}

type closerFunc func()

func (f closerFunc) Close() error {
	f()
	return nil
}

func TestTLS(t *testing.T) {
	times := 1000
	idMap := map[int64]unsafe.Pointer{}
	idMapMu := sync.Mutex{}

	for i := 0; i < times; i++ {
		t.Run(fmt.Sprintf("Round %v", i), func(t *testing.T) {
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

			// fmt.Println(g.G(), GoID(), tls.ID())
			id := gls.ID()

			if id <= 0 {
				t.Fatalf("fail to get ID. [id:%v]", id)
			}

			idMapMu.Lock()
			defer idMapMu.Unlock()

			if p, ok := idMap[id]; ok {
				t.Fatalf("duplicated ID. [id:%v], %p, %p", id, p, g.G())
			}

			id = gls.ID()

			idMap[id] = g.G()
		})
	}
}

func TestUnload(t *testing.T) {
	// Run test in a standalone goroutine.
	t.Run("try unload", func(t *testing.T) {
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

		if _, ok := gls.Get(key); ok {
			t.Fatalf("key must be cleared. [key:%v]", key)
		}

		if exitCalled {
			t.Fatalf("all AtExit functions must not be called.")
		}
	})

	t.Run("try Reload", func(t *testing.T) {
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

		if _, ok := gls.Get(key); ok {
			t.Fatalf("key must be cleared. [key:%v]", key)
		}

		if exitCalled {
			t.Fatalf("all AtExit functions must not be called.")
		}

		gls.Set(key, gls.MakeData(expected))
		if d, ok := gls.Get(key); !ok {
			t.Fatalf("fail to get data. [key:%v]", key)
		} else if actual, ok := d.Value().(string); !ok || actual != expected {
			t.Fatalf("invalid value. [key:%v] [value:%v] [expected:%v]", key, actual, expected)
		}

	})
}

func TestShrinkStack(t *testing.T) {
	const times = 1000
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
		case <-time.After(10 * time.Second):
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
	time.Sleep(time.Millisecond * 100)

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

type key struct{}

// BenchmarkSetGLS/normal_set-10         	20790284	        57.46 ns/op	      40 B/op	       1 allocs/op
// BenchmarkSetGLS/normal_get-10         	77068196	        15.45 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSetGLS/set_kv-10             	42193660	        28.15 ns/op	       8 B/op	       0 allocs/op
// BenchmarkSetGLS/get_kv-10             	79144359	        15.06 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSetGLS/set_gls-10            	90627027	        13.00 ns/op	       8 B/op	       0 allocs/op
// BenchmarkSetGLS/get_gls-10            	283466686	        4.242 ns/op	       0 B/op	       0 allocs/opp
func BenchmarkSetGLS(b *testing.B) {
	b.Run("normal set", func(b *testing.B) {
		b.ReportAllocs()
		gls.SetKV(key{}, 0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gls.Set(key{}, gls.MakeData(i))
		}
	})

	b.Run("normal get", func(b *testing.B) {
		b.ReportAllocs()
		gls.Set(key{}, gls.MakeData(0))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if v, ok := gls.Get(key{}); ok {
				if v2, ok := v.Value().(int); ok {
					_ = v2
				}
			}
		}
	})

	b.Run("set kv", func(b *testing.B) {
		b.ReportAllocs()
		gls.SetKV(key{}, 0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gls.SetKV(key{}, i)
		}
	})

	b.Run("get kv", func(b *testing.B) {
		b.ReportAllocs()
		gls.SetKV(key{}, 0)
		var x int
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gls.GetKV[int](key{})
		}
		_ = x
	})

	b.Run("set gls", func(b *testing.B) {
		b.ReportAllocs()
		gls.SetGLS(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gls.SetGLS(i)
		}
	})

	b.Run("get gls", func(b *testing.B) {
		b.ReportAllocs()
		gls.SetGLS(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = gls.GetGLS[int]()
		}
	})
}
