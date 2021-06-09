package carray

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"github.com/funbytes/modern-go/internal/rwmutex"
)

// Array is a golang array with rich features.
// It contains a concurrent-safe/unsafe switch, which should be set
// when its initialization and cannot be changed then.
type Array struct {
	mu    *rwmutex.RWMutex
	array []interface{}
}

// New creates and returns an empty array.
// The parameter <safe> is used to specify whether using array in concurrent-safety,
// which is false in default.
func New(safe ...bool) *Array {
	return NewArraySize(0, 0, safe...)
}

// NewArraySize create and returns an array with given size and cap.
// The parameter <safe> is used to specify whether using array in concurrent-safety,
// which is false in default.
func NewArraySize(size int, cap int, safe ...bool) *Array {
	return &Array{
		mu:    rwmutex.New(safe...),
		array: make([]interface{}, size, cap),
	}
}

// NewArrayFrom creates and returns an array with given slice <array>.
// The parameter <safe> is used to specify whether using array in concurrent-safety,
// which is false in default.
func NewArrayFrom(array []interface{}, safe ...bool) *Array {
	return &Array{
		mu:    rwmutex.New(safe...),
		array: array,
	}
}

// NewArrayFromCopy creates and returns an array from a copy of given slice <array>.
// The parameter <safe> is used to specify whether using array in concurrent-safety,
// which is false in default.
func NewArrayFromCopy(array []interface{}, safe ...bool) *Array {
	newArray := make([]interface{}, len(array))
	copy(newArray, array)
	return &Array{
		mu:    rwmutex.New(safe...),
		array: newArray,
	}
}

// Get returns the value by the specified index.
// If the given <index> is out of range of the array, the <found> is false.
func (a *Array) Get(index int) (value interface{}, found bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if index < 0 || index >= len(a.array) {
		return nil, false
	}
	return a.array[index], true
}

// Set sets value to specified index.
func (a *Array) Set(index int, value interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index < 0 || index >= len(a.array) {
		return errors.New(fmt.Sprintf("index %d out of array range %d", index, len(a.array)))
	}
	a.array[index] = value
	return nil
}

// SetArray sets the underlying slice array with the given <array>.
func (a *Array) SetArray(array []interface{}) *Array {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.array = array
	return a
}

// Replace replaces the array items by given <array> from the beginning of array.
func (a *Array) Replace(array []interface{}) *Array {
	a.mu.Lock()
	defer a.mu.Unlock()
	max := len(array)
	if max > len(a.array) {
		max = len(a.array)
	}
	for i := 0; i < max; i++ {
		a.array[i] = array[i]
	}
	return a
}

// SortFunc sorts the array by custom function <less>.
func (a *Array) SortFunc(less func(v1, v2 interface{}) bool) *Array {
	a.mu.Lock()
	defer a.mu.Unlock()
	sort.Slice(a.array, func(i, j int) bool {
		return less(a.array[i], a.array[j])
	})
	return a
}

// InsertBefore inserts the <value> to the front of <index>.
func (a *Array) InsertBefore(index int, value interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index < 0 || index >= len(a.array) {
		return errors.New(fmt.Sprintf("index %d out of array range %d", index, len(a.array)))
	}
	rear := append([]interface{}{}, a.array[index:]...)
	a.array = append(a.array[0:index], value)
	a.array = append(a.array, rear...)
	return nil
}

// InsertAfter inserts the <value> to the back of <index>.
func (a *Array) InsertAfter(index int, value interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index < 0 || index >= len(a.array) {
		return errors.New(fmt.Sprintf("index %d out of array range %d", index, len(a.array)))
	}
	rear := append([]interface{}{}, a.array[index+1:]...)
	a.array = append(a.array[0:index+1], value)
	a.array = append(a.array, rear...)
	return nil
}

// Remove removes an item by index.
// If the given <index> is out of range of the array, the <found> is false.
func (a *Array) Remove(index int) (value interface{}, found bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.doRemoveWithoutLock(index)
}

// doRemoveWithoutLock removes an item by index without lock.
func (a *Array) doRemoveWithoutLock(index int) (value interface{}, found bool) {
	if index < 0 || index >= len(a.array) {
		return nil, false
	}
	// Determine array boundaries when deleting to improve deletion efficiency.
	if index == 0 {
		value := a.array[0]
		a.array = a.array[1:]
		return value, true
	} else if index == len(a.array)-1 {
		value := a.array[index]
		a.array = a.array[:index]
		return value, true
	}
	// If it is a non-boundary delete,
	// it will involve the creation of an array,
	// then the deletion is less efficient.
	value = a.array[index]
	a.array = append(a.array[:index], a.array[index+1:]...)
	return value, true
}

// RemoveValue removes an item by value.
// It returns true if value is found in the array, or else false if not found.
func (a *Array) RemoveValue(value interface{}) bool {
	if i := a.Search(value); i != -1 {
		a.Remove(i)
		return true
	}
	return false
}

// PushLeft pushes one or multiple items to the beginning of array.
func (a *Array) PushLeft(value ...interface{}) *Array {
	a.mu.Lock()
	a.array = append(value, a.array...)
	a.mu.Unlock()
	return a
}

// PushRight pushes one or multiple items to the end of array.
// It equals to Append.
func (a *Array) PushRight(value ...interface{}) *Array {
	a.mu.Lock()
	a.array = append(a.array, value...)
	a.mu.Unlock()
	return a
}

// PopRand randomly pops and return an item out of array.
// Note that if the array is empty, the <found> is false.
func (a *Array) PopRand() (value interface{}, found bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.doRemoveWithoutLock(rand.Intn(len(a.array)))
}

// PopRands randomly pops and returns <size> items out of array.
func (a *Array) PopRands(size int) []interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	if size <= 0 || len(a.array) == 0 {
		return nil
	}
	if size >= len(a.array) {
		size = len(a.array)
	}
	array := make([]interface{}, size)
	for i := 0; i < size; i++ {
		array[i], _ = a.doRemoveWithoutLock(rand.Intn(len(a.array)))
	}
	return array
}

// PopLeft pops and returns an item from the beginning of array.
// Note that if the array is empty, the <found> is false.
func (a *Array) PopLeft() (value interface{}, found bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.array) == 0 {
		return nil, false
	}
	value = a.array[0]
	a.array = a.array[1:]
	return value, true
}

// PopRight pops and returns an item from the end of array.
// Note that if the array is empty, the <found> is false.
func (a *Array) PopRight() (value interface{}, found bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	index := len(a.array) - 1
	if index < 0 {
		return nil, false
	}
	value = a.array[index]
	a.array = a.array[:index]
	return value, true
}

// PopLefts pops and returns <size> items from the beginning of array.
func (a *Array) PopLefts(size int) []interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	if size <= 0 || len(a.array) == 0 {
		return nil
	}
	if size >= len(a.array) {
		array := a.array
		a.array = a.array[:0]
		return array
	}
	value := a.array[0:size]
	a.array = a.array[size:]
	return value
}

// PopRights pops and returns <size> items from the end of array.
func (a *Array) PopRights(size int) []interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	if size <= 0 || len(a.array) == 0 {
		return nil
	}
	index := len(a.array) - size
	if index <= 0 {
		array := a.array
		a.array = a.array[:0]
		return array
	}
	value := a.array[index:]
	a.array = a.array[:index]
	return value
}

// Range picks and returns items by range, like array[start:end].
// Notice, if in concurrent-safe usage, it returns a copy of slice;
// else a pointer to the underlying data.
//
// If <end> is negative, then the offset will start from the end of array.
// If <end> is omitted, then the sequence will have everything from start up
// until the end of the array.
func (a *Array) Range(start int, end ...int) []interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	offsetEnd := len(a.array)
	if len(end) > 0 && end[0] < offsetEnd {
		offsetEnd = end[0]
	}
	if start > offsetEnd {
		return nil
	}
	if start < 0 {
		start = 0
	}
	array := ([]interface{})(nil)
	if a.mu.IsSafe() {
		array = make([]interface{}, offsetEnd-start)
		copy(array, a.array[start:offsetEnd])
	} else {
		array = a.array[start:offsetEnd]
	}
	return array
}

// See PushRight.
func (a *Array) Append(value ...interface{}) *Array {
	a.PushRight(value...)
	return a
}

// Len returns the length of array.
func (a *Array) Len() int {
	a.mu.RLock()
	length := len(a.array)
	a.mu.RUnlock()
	return length
}

// Clone returns a new array, which is a copy of current array.
func (a *Array) Clone() (newArray *Array) {
	a.mu.RLock()
	array := make([]interface{}, len(a.array))
	copy(array, a.array)
	a.mu.RUnlock()
	return NewArrayFrom(array, a.mu.IsSafe())
}

// Clear deletes all items of current array.
func (a *Array) Clear() *Array {
	a.mu.Lock()
	if len(a.array) > 0 {
		a.array = make([]interface{}, 0)
	}
	a.mu.Unlock()
	return a
}

// Contains checks whether a value exists in the array.
func (a *Array) Contains(value interface{}) bool {
	return a.Search(value) != -1
}

// Search searches array by <value>, returns the index of <value>,
// or returns -1 if not exists.
func (a *Array) Search(value interface{}) int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.array) == 0 {
		return -1
	}
	result := -1
	for index, v := range a.array {
		if v == value {
			result = index
			break
		}
	}
	return result
}

// Unique uniques the array, clear repeated items.
// Example: [1,1,2,3,2] -> [1,2,3]
func (a *Array) Unique() *Array {
	a.mu.Lock()
	for i := 0; i < len(a.array)-1; i++ {
		for j := i + 1; j < len(a.array); {
			if a.array[i] == a.array[j] {
				a.array = append(a.array[:j], a.array[j+1:]...)
			} else {
				j++
			}
		}
	}
	a.mu.Unlock()
	return a
}

// LockFunc locks writing by callback function <f>.
func (a *Array) LockFunc(f func(array []interface{})) *Array {
	a.mu.Lock()
	defer a.mu.Unlock()
	f(a.array)
	return a
}

// RLockFunc locks reading by callback function <f>.
func (a *Array) RLockFunc(f func(array []interface{})) *Array {
	a.mu.RLock()
	defer a.mu.RUnlock()
	f(a.array)
	return a
}

// Rand randomly returns one item from array(no deleting).
func (a *Array) Rand() (value interface{}, found bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.array) == 0 {
		return nil, false
	}
	return a.array[rand.Intn(len(a.array))], true
}

// Rands randomly returns <size> items from array(no deleting).
func (a *Array) Rands(size int) []interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if size <= 0 || len(a.array) == 0 {
		return nil
	}
	array := make([]interface{}, size)
	for i := 0; i < size; i++ {
		array[i] = a.array[rand.Intn(len(a.array))]
	}
	return array
}

// Shuffle randomly shuffles the array.
func (a *Array) Shuffle() *Array {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, v := range rand.Perm(len(a.array)) {
		a.array[i], a.array[v] = a.array[v], a.array[i]
	}
	return a
}

// Reverse makes array with elements in reverse order.
func (a *Array) Reverse() *Array {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, j := 0, len(a.array)-1; i < j; i, j = i+1, j-1 {
		a.array[i], a.array[j] = a.array[j], a.array[i]
	}
	return a
}

// CountValues counts the number of occurrences of all values in the array.
func (a *Array) CountValues() map[interface{}]int {
	m := make(map[interface{}]int)
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, v := range a.array {
		m[v]++
	}
	return m
}

// Iterator is alias of IteratorAsc.
func (a *Array) Iterator(f func(k int, v interface{}) bool) {
	a.IteratorAsc(f)
}

// IteratorAsc iterates the array readonly in ascending order with given callback function <f>.
// If <f> returns true, then it continues iterating; or false to stop.
func (a *Array) IteratorAsc(f func(k int, v interface{}) bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for k, v := range a.array {
		if !f(k, v) {
			break
		}
	}
}

// IteratorDesc iterates the array readonly in descending order with given callback function <f>.
// If <f> returns true, then it continues iterating; or false to stop.
func (a *Array) IteratorDesc(f func(k int, v interface{}) bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for i := len(a.array) - 1; i >= 0; i-- {
		if !f(i, a.array[i]) {
			break
		}
	}
}

// Walk applies a user supplied function <f> to every item of array.
func (a *Array) Walk(f func(value interface{}) interface{}) *Array {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, v := range a.array {
		a.array[i] = f(v)
	}
	return a
}

// IsEmpty checks whether the array is empty.
func (a *Array) IsEmpty() bool {
	return a.Len() == 0
}
