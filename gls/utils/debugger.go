//go:build debugger

package utils

// AddGoroutineTag 直接打印 tag: $filename/$funcname()
func AddGoroutineTag() {
	SetLabels(func() []string {
		return []string{
			"function", GetFileAndFuncName(),
		}
	})
}

// AddGoroutineTagParams 打印 tag: $filename/$funcname()
// 打印 额外的参数， 需要是偶数个参数
func AddGoroutineTagParams(KeyValue ...string) {
	SetLabels(func() []string {
		var list = []string{
			"function", GetFileAndFuncName(),
		}
		if len(KeyValue)%2 != 0 {
			KeyValue = append(KeyValue, "")
		}
		for i := len(KeyValue) - 2; i >= 0; i -= 2 {
			list = append(list, KeyValue[i])
			list = append(list, KeyValue[i+1])
		}
		return list
	})
}
