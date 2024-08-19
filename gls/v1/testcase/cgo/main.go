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

	v1 "github.com/funbytes/modern-go/gls/v1"
)

func main() {
	C.printint(C.int(44))

	go func() {
		v1.AtExit(func() {
			fmt.Printf("goroutine:%d exit\n", v1.ID())
		})

		fmt.Printf("goroutine:%d start\n", v1.ID())
		v1.Set("kk1", v1.MakeData("ccc"))
		cc, _ := v1.Get("kk1")
		fmt.Printf("%v\n", cc.Value().(string))
	}()

	time.Sleep(5 * time.Second)
	fmt.Printf("main exit")
}
