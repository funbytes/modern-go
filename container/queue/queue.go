package queue

import (
	"container/list"
	"sync"
)

type Queue struct {
	list *list.List
	lock *sync.RWMutex
}

// 新建一个队列
func NewQueue() *Queue {
	this := new(Queue)
	this.list = list.New()
	this.lock = new(sync.RWMutex)
	return this
}

// 返回队列长度
func (q *Queue) Length() int {
	q.lock.RLock()
	l := q.list.Len()
	q.lock.RUnlock()
	return l
}

// 返回队列头
func (q *Queue) Front() interface{} {
	q.lock.RLock()
	e := q.list.Front()
	if e != nil {
		q.lock.RUnlock()
		return e.Value
	}
	q.lock.RUnlock()
	return nil
}

// 队列尾部插入元素
func (q *Queue) Push(v interface{}) *list.Element {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.list.PushBack(v)

}

func (q *Queue) Remove(e *list.Element) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if e != nil {
		q.list.Remove(e)
	}
}

// 队列头部弹出元素
func (q *Queue) Pop() interface{} {
	q.lock.Lock()
	e := q.list.Front()
	if e != nil {
		value := e.Value
		q.list.Remove(e)
		q.lock.Unlock()
		return value
	}
	q.lock.Unlock()
	return nil
}

// 移动到队尾
func (q *Queue) MoveBack(e *list.Element) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.list.MoveToBack(e)
}

// 清空队列
func (q *Queue) Clear() {
	q.lock.Lock()
	q.list = q.list.Init()
	q.lock.Unlock()
}
