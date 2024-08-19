package tls

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"testing"
	"time"

	"gitlab-ee.funplus.io/watcher/watcher/misc/godebugtag"
	"gitlab-ee.funplus.io/watcher/watcher/misc/gotls/g"
)

// TestSetLabels 测试SetLabels的race检测
// 1. 打开-race运行（记得改下面的超时时间）
// 2. 打开链接：http://localhost:9788/debug/pprof/goroutine?debug=1
// 3. 看控制台是否会报race警告
func TestSetLabels(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.SetOutput(os.Stdout)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() {
		if err := http.ListenAndServe(":9788", nil); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	godebugtag.SetLabels(func() []string {
		return []string{"1", "2"}
	})
	AtExit(func() {})
	v := routineTimer(g.G())
	t.Log("--->", v)
	for i := 0; i < 5; i++ {
		d := i
		go func(index int) {
			godebugtag.AddGoroutineTag()
			v := routineTimer(g.G())
			t.Log("--->", d, v)
			AtExit(func() {})
			v = routineTimer(g.G())
			t.Log("--+>", d, v)
			<-ctx.Done()
		}(i)
	}

	<-ctx.Done()
}
