package syncmap

import "sync"

func Length(m *sync.Map) int {
	if m == nil {
		return 0
	}
	length := 0
	m.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}
