package stack

import (
	"container/list"
	"sync"
)

type Stack struct {
	list *list.List
	lock *sync.RWMutex
}

//新建一个栈
func NewStack() *Stack {
	this := new(Stack)
	this.list = list.New()
	this.lock = new(sync.RWMutex)
	return this
}

//返回栈长度
func (this *Stack) Length() int {
	this.lock.RLock()
	l := this.list.Len()
	this.lock.RUnlock()
	return l
}

//栈尾部插入元素
func (this *Stack) Push(v interface{}) {
	this.lock.Lock()
	this.list.PushBack(v)
	this.lock.Unlock()
}

//栈尾部弹出元素
func (this *Stack) Pop() interface{} {
	this.lock.Lock()
	e := this.list.Back()
	if e != nil {
		value := e.Value
		this.list.Remove(e)
		this.lock.Unlock()
		return value
	}
	this.lock.Unlock()
	return nil
}

//清空栈
func (this *Stack) Clear() {
	this.lock.Lock()
	this.list = this.list.Init()
	this.lock.Unlock()
}
