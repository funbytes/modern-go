package cmap

import (
	"sync"
)

// A thread safe map.
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards.
var (
	SHARD_COUNT = 32
)

// template type ConcurrentMap(KType,VType,KeyHash)
type KType string
type VType interface{}

type ConcurrentMap []*sharded

type sharded struct {
	items map[KType]VType
	sync.RWMutex
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type Tuple struct {
	Key KType
	Val VType
}

func New() ConcurrentMap {
	this := make(ConcurrentMap, SHARD_COUNT)
	for i := 0; i < SHARD_COUNT; i++ {
		this[i] = &sharded{items: make(map[KType]VType)}
	}
	return this
}

func KeyHash(key KType) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

// Returns shard under given key.
func (m ConcurrentMap) GetShard(key KType) *sharded {
	return m[uint64(KeyHash(key))%uint64(SHARD_COUNT)]
}

// IsEmpty checks if map is empty.
func (m ConcurrentMap) IsEmpty() bool {
	return m.Count() == 0
}

func (m *ConcurrentMap) Set(key KType, value VType) {
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

// get all keys
func (m ConcurrentMap) Keys() []KType {
	var ret []KType
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
func (m ConcurrentMap) MGet(keys ...KType) map[KType]VType {
	data := make(map[KType]VType)
	for _, key := range keys {
		if val, ok := m.Get(key); ok {
			data[key] = val
		}
	}
	return data
}

// get all values
func (m ConcurrentMap) GetAll() map[KType]VType {
	data := make(map[KType]VType)

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
func (m ConcurrentMap) Clear() {
	for _, shard := range m {
		shard.Lock()
		shard.items = make(map[KType]VType)
		shard.Unlock()
	}
}

// multiple set
func (m *ConcurrentMap) MSet(data map[KType]VType) {
	for key, value := range data {
		m.Set(key, value)
	}
}

// like redis SETNX
// return true if the key was set
// return false if the key was not set
func (m *ConcurrentMap) SetNX(key KType, value VType) bool {
	shard := m.GetShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return true
}

func (m ConcurrentMap) Get(key KType) (VType, bool) {
	shard := m.GetShard(key)
	shard.RLock()
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

func (m ConcurrentMap) Count() int {
	count := 0
	for i := 0; i < SHARD_COUNT; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

func (m *ConcurrentMap) Has(key KType) bool {
	shard := m.GetShard(key)
	shard.RLock()
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

func (m *ConcurrentMap) Remove(key KType) {
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

func (m ConcurrentMap) GetAndRemove(key KType) (VType, bool) {
	shard := m.GetShard(key)
	shard.Lock()
	val, ok := shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return val, ok
}

// Returns an iterator which could be used in a for range loop.
func (m ConcurrentMap) Iter() <-chan Tuple {
	ch := make(chan Tuple)
	go func() {
		for _, shard := range m {
			shard.RLock()
			for key, val := range shard.items {
				ch <- Tuple{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

// Returns a buffered iterator which could be used in a for range loop.
func (m ConcurrentMap) IterBuffered() <-chan Tuple {
	ch := make(chan Tuple, m.Count())
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- Tuple{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}
