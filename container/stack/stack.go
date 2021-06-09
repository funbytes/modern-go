package stack

import (
	"container/list"
	"sync"
)

type Stack struct {
	list *list.List
	lock *sync.RWMutex
}

// 新建一个栈
func NewStack() *Stack {
	this := new(Stack)
	this.list = list.New()
	this.lock = new(sync.RWMutex)
	return this
}

// 返回栈长度
func (s *Stack) Length() int {
	s.lock.RLock()
	l := s.list.Len()
	s.lock.RUnlock()
	return l
}

// 栈尾部插入元素
func (s *Stack) Push(v interface{}) {
	s.lock.Lock()
	s.list.PushBack(v)
	s.lock.Unlock()
}

// 栈尾部弹出元素
func (s *Stack) Pop() interface{} {
	s.lock.Lock()
	e := s.list.Back()
	if e != nil {
		value := e.Value
		s.list.Remove(e)
		s.lock.Unlock()
		return value
	}
	s.lock.Unlock()
	return nil
}

// 清空栈
func (s *Stack) Clear() {
	s.lock.Lock()
	s.list = s.list.Init()
	s.lock.Unlock()
}
