// Code generated by gotemplate. DO NOT EDIT.

package cmap

import (
	"sync"
)

// A thread safe map.
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards.
var (
	SHARD_COUNTConcurrentMapUint32Uint64 = 32
)

// template type ConcurrentMap(KType,VType,KeyHash)

type ConcurrentMapUint32Uint64 []*shardedConcurrentMapUint32Uint64

type shardedConcurrentMapUint32Uint64 struct {
	items map[uint32]uint64
	sync.RWMutex
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type TupleConcurrentMapUint32Uint64 struct {
	Key uint32
	Val uint64
}

func NewConcurrentMapUint32Uint64() ConcurrentMapUint32Uint64 {
	this := make(ConcurrentMapUint32Uint64, SHARD_COUNTConcurrentMapUint32Uint64)
	for i := 0; i < SHARD_COUNTConcurrentMapUint32Uint64; i++ {
		this[i] = &shardedConcurrentMapUint32Uint64{items: make(map[uint32]uint64)}
	}
	return this
}

// Returns shard under given key.
func (m ConcurrentMapUint32Uint64) GetShard(key uint32) *shardedConcurrentMapUint32Uint64 {
	return m[uint64(KeyHashUint32(key))%uint64(SHARD_COUNTConcurrentMapUint32Uint64)]
}

// IsEmpty checks if map is empty.
func (m ConcurrentMapUint32Uint64) IsEmpty() bool {
	return m.Count() == 0
}

func (m *ConcurrentMapUint32Uint64) Set(key uint32, value uint64) {
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

// get all keys
func (m ConcurrentMapUint32Uint64) Keys() []uint32 {
	var ret []uint32
	for _, shard := range m {
		shard.RLock()
		for key := range shard.items {
			ret = append(ret, key)
		}
		shard.RUnlock()
	}
	return ret
}

// multiple get by keys
func (m ConcurrentMapUint32Uint64) MGet(keys ...uint32) map[uint32]uint64 {
	data := make(map[uint32]uint64)
	for _, key := range keys {
		if val, ok := m.Get(key); ok {
			data[key] = val
		}
	}
	return data
}

// get all values
func (m ConcurrentMapUint32Uint64) GetAll() map[uint32]uint64 {
	data := make(map[uint32]uint64)

	for _, shard := range m {
		shard.RLock()
		for key, val := range shard.items {
			data[key] = val
		}
		shard.RUnlock()
	}
	return data
}

// clear all values
func (m ConcurrentMapUint32Uint64) Clear() {
	for _, shard := range m {
		shard.Lock()
		shard.items = make(map[uint32]uint64)
		shard.Unlock()
	}
}

// multiple set
func (m *ConcurrentMapUint32Uint64) MSet(data map[uint32]uint64) {
	for key, value := range data {
		m.Set(key, value)
	}
}

// like redis SETNX
// return true if the key was set
// return false if the key was not set
func (m *ConcurrentMapUint32Uint64) SetNX(key uint32, value uint64) bool {
	shard := m.GetShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return true
}

func (m ConcurrentMapUint32Uint64) Get(key uint32) (uint64, bool) {
	shard := m.GetShard(key)
	shard.RLock()
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

func (m ConcurrentMapUint32Uint64) Count() int {
	count := 0
	for i := 0; i < SHARD_COUNTConcurrentMapUint32Uint64; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

func (m *ConcurrentMapUint32Uint64) Has(key uint32) bool {
	shard := m.GetShard(key)
	shard.RLock()
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

func (m *ConcurrentMapUint32Uint64) Remove(key uint32) {
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

func (m ConcurrentMapUint32Uint64) GetAndRemove(key uint32) (uint64, bool) {
	shard := m.GetShard(key)
	shard.Lock()
	val, ok := shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return val, ok
}

// Returns an iterator which could be used in a for range loop.
func (m ConcurrentMapUint32Uint64) Iter() <-chan TupleConcurrentMapUint32Uint64 {
	ch := make(chan TupleConcurrentMapUint32Uint64)
	go func() {
		for _, shard := range m {
			shard.RLock()
			for key, val := range shard.items {
				ch <- TupleConcurrentMapUint32Uint64{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

// Returns a buffered iterator which could be used in a for range loop.
func (m ConcurrentMapUint32Uint64) IterBuffered() <-chan TupleConcurrentMapUint32Uint64 {
	ch := make(chan TupleConcurrentMapUint32Uint64, m.Count())
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- TupleConcurrentMapUint32Uint64{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}
