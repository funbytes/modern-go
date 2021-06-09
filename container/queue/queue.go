package queue

import (
	"container/list"
	"sync"
)

type Queue struct {
	list *list.List
	lock *sync.RWMutex
}

//新建一个队列
func NewQueue() *Queue {
	this := new(Queue)
	this.list = list.New()
	this.lock = new(sync.RWMutex)
	return this
}

//返回队列长度
func (this *Queue) Length() int {
	this.lock.RLock()
	l := this.list.Len()
	this.lock.RUnlock()
	return l
}

//返回队列头
func (this *Queue) Front() interface{} {
	this.lock.RLock()
	e := this.list.Front()
	if e != nil {
		this.lock.RUnlock()
		return e.Value
	}
	this.lock.RUnlock()
	return nil
}

//队列尾部插入元素
func (this *Queue) Push(v interface{}) *list.Element {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.list.PushBack(v)

}

func (this *Queue) Remove(e *list.Element) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if e != nil {
		this.list.Remove(e)
	}
}

//队列头部弹出元素
func (this *Queue) Pop() interface{} {
	this.lock.Lock()
	e := this.list.Front()
	if e != nil {
		value := e.Value
		this.list.Remove(e)
		this.lock.Unlock()
		return value
	}
	this.lock.Unlock()
	return nil
}

//移动到队尾
func (this *Queue) MoveBack(e *list.Element) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.list.MoveToBack(e)
}

//清空队列
func (this *Queue) Clear() {
	this.lock.Lock()
	this.list = this.list.Init()
	this.lock.Unlock()
}
