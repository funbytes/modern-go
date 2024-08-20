//go:build !debugger
// +build !debugger

package utils

// AddGoroutineTag 直接打印 tag: $filename/$funcname()
func AddGoroutineTag() {}

// AddGoroutineTagParams 打印 tag: $filename/$funcname()
// 打印 额外的参数， 需要是偶数个参数
func AddGoroutineTagParams(KeyValue ...string) {}
