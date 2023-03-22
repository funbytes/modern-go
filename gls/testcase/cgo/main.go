package main

/*
#include <stdio.h>

void printint(int v) {
    printf("printint: %d\n", v);
}
*/
import "C"

import (
	"fmt"
	"time"

	"github.com/funbytes/modern-go/gls"
)

func main() {
	C.printint(C.int(44))

	go func() {
		gls.AtExit(func() {
			fmt.Printf("goroutine:%d exit\n", gls.ID())
		})

		fmt.Printf("goroutine:%d start\n", gls.ID())
		gls.Set("kk1", gls.MakeData("ccc"))
		cc, _ := gls.Get("kk1")
		fmt.Printf("%v\n", cc.Value().(string))
	}()

	time.Sleep(5 * time.Second)
	fmt.Printf("main exit")
}
