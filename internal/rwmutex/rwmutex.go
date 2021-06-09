// Package rwmutex provides switch of concurrent safety feature for sync.RWMutex.
package rwmutex

import "sync"

// RWMutex is a sync.RWMutex with a switch for concurrent safe feature.
type RWMutex struct {
	*sync.RWMutex
}

// New creates and returns a new *RWMutex.
func New(safe ...bool) *RWMutex {
	mu := RWMutex{}
	if len(safe) > 0 && safe[0] {
		mu.RWMutex = new(sync.RWMutex)
	}
	return &mu
}

// IsSafe checks and returns whether current mutex is in concurrent-safe usage.
func (mu *RWMutex) IsSafe() bool {
	return mu.RWMutex != nil
}

// Lock locks mutex for writing.
// It does nothing if it is not in concurrent-safe usage.
func (mu *RWMutex) Lock() {
	if mu.RWMutex != nil {
		mu.RWMutex.Lock()
	}
}

// Unlock unlocks mutex for writing.
// It does nothing if it is not in concurrent-safe usage.
func (mu *RWMutex) Unlock() {
	if mu.RWMutex != nil {
		mu.RWMutex.Unlock()
	}
}

// RLock locks mutex for reading.
// It does nothing if it is not in concurrent-safe usage.
func (mu *RWMutex) RLock() {
	if mu.RWMutex != nil {
		mu.RWMutex.RLock()
	}
}

// RUnlock unlocks mutex for reading.
// It does nothing if it is not in concurrent-safe usage.
func (mu *RWMutex) RUnlock() {
	if mu.RWMutex != nil {
		mu.RWMutex.RUnlock()
	}
}
