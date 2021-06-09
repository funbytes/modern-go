package skiplist

import (
	"errors"
	//"log"
	"math/rand"
	"sync"
	"time"
)

const DefaultMaxLevel = 32
const p = 0.25

type node struct {
	key      string
	score    float64
	backward *node
	level    [DefaultMaxLevel]struct {
		forward *node
		span    int64
	}
}

type Node struct {
	Key   string  `json:"key"`
	Score float64 `json:"score"`
}

type skipList struct {
	header *node
	tail   *node
	length int64
	level  int
}

type Set struct {
	dict     map[string]*float64
	skipList *skipList
	lock     *sync.RWMutex
}

func newSkipList() *skipList {
	sl := new(skipList)
	sl.level = 1
	sl.length = 0
	sl.header = new(node)
	for j := 0; j < DefaultMaxLevel; j++ {
		sl.header.level[j].forward = nil
		sl.header.level[j].span = 0
	}
	sl.header.backward = nil
	sl.tail = nil
	return sl
}

func NewSet() *Set {
	sl := newSkipList()
	sl.level = 1
	return &Set{
		dict:     make(map[string]*float64),
		skipList: sl,
		lock:     new(sync.RWMutex),
	}
}

func (this *Set) createNode(key string, score float64) *node {
	n := new(node)
	n.key = key
	n.score = score
	return n
}

//根据排名获取node 按升序获得 rank 从1开始
func (this *Set) GetElementByRankASC(rank int64) Node {
	this.lock.RLock()
	defer this.lock.RUnlock()
	var n Node
	x := this.skipList.header
	var traversed int64 = 0
	for i := this.skipList.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span) <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if traversed == rank {
			n.Key = x.key
			n.Score = x.score
			return n
		}
	}
	return n
}

//根据排名获取node 按降序获得 rank 从1开始
func (this *Set) GetElementByRankDESC(rank int64) Node {
	rank = this.GetLenth() - rank + 1
	return this.GetElementByRankASC(rank)
}
func (this *Set) GetScore(key string) (float64, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if _, ok := this.dict[key]; !ok {
		return 0, errors.New("not find key " + key)
	}
	score := *this.dict[key]
	return score, nil
}

//按升序排名 从1开始
func (this *Set) GetRank(key string) (int64, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if _, ok := this.dict[key]; !ok {
		return 0, errors.New("not find key " + key)
	}
	var x *node
	var rank int64 = 0
	score := *this.dict[key]
	x = this.skipList.header
	for i := this.skipList.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.key <= key)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}
		if x.key == key {
			return rank, nil
		}
	}
	return 0, nil
}

//按降序排名
func (this *Set) GetRankDESC(key string) (int64, error) {
	rank, err := this.GetRank(key)
	if err != nil {
		return 0, err
	}
	return this.GetLenth() - rank + 1, nil
}

func (this *Set) Del(key string) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.dict[key]; !ok {
		return errors.New("not find key " + key)
	}
	score := *this.dict[key]
	var update [DefaultMaxLevel]*node
	var x *node
	x = this.skipList.header
	var i int
	for i = this.skipList.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.key < key)) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	x = x.level[0].forward
	if x != nil && score == x.score && key == x.key {
		for i = 0; i < this.skipList.level; i++ {
			if update[i].level[i].forward == x {
				update[i].level[i].span += x.level[i].span - 1
				update[i].level[i].forward = x.level[i].forward
			} else {
				update[i].level[i].span--
			}
		}
		if x.level[0].forward != nil {
			x.level[0].forward.backward = x.backward
		} else {
			this.skipList.tail = x.backward
		}
		for this.skipList.level > 1 && this.skipList.header.level[this.skipList.level-1].forward == nil {
			this.skipList.level--
		}
		this.skipList.length--
	}
	delete(this.dict, key)
	return nil
}

func (this *Set) Set(key string, score float64) *node {
	this.Del(key)
	node := this.insert(key, score)
	return node
}

//获取top n个数 按升序获取
func (this *Set) GetTopN(n int64) []Node {
	this.lock.RLock()
	defer this.lock.RUnlock()
	var nodes []Node
	var q int64 = 0
	for x := this.skipList.header.level[0].forward; x != nil && q < n; x = x.level[0].forward {
		var node Node
		node.Key = x.key
		node.Score = x.score
		nodes = append(nodes, node)
		q++
	}
	return nodes
}

//获取top n个数 按降序获取
func (this *Set) GetTopNDESC(n int64) []Node {
	this.lock.RLock()
	defer this.lock.RUnlock()
	var nodes []Node
	var q int64 = 0
	for x := this.skipList.tail; x != nil && q < n; x = x.backward {
		var node Node
		node.Key = x.key
		node.Score = x.score
		nodes = append(nodes, node)
		q++
	}
	return nodes
}

func (this *Set) insert(key string, score float64) *node {
	this.lock.Lock()
	defer this.lock.Unlock()
	var update [DefaultMaxLevel]*node
	var x *node
	var rank [DefaultMaxLevel]int64
	var i, level int

	x = this.skipList.header

	for i = this.skipList.level - 1; i >= 0; i-- {
		rank[i] = 0
		if i != this.skipList.level-1 {
			rank[i] = rank[i+1]
		}
		for x.level[i].forward != nil && (x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.key < key)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}
	level = this.randomLevel()
	if level > this.skipList.level {
		for i = this.skipList.level; i < level; i++ {
			rank[i] = 0
			update[i] = this.skipList.header
			update[i].level[i].span = this.skipList.length
		}
		this.skipList.level = level
	}
	x = this.createNode(key, score)
	for i = 0; i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}
	for i = level; i < this.skipList.level; i++ {
		update[i].level[i].span++
	}
	if update[0] != this.skipList.header {
		x.backward = update[0]
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		this.skipList.tail = x
	}
	this.skipList.length++
	this.dict[key] = &x.score
	return x
}

func (this *Set) GetLenth() int64 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.skipList.length
}

func (this *Set) GetLevel() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.skipList.level
}

func (this *Set) randomLevel() int {
	rand.Seed(time.Now().UnixNano())
	p := 0.25
	level := 1
	for (float64(rand.Int63() & 0xFFFF)) < (p * 0xFFFF) {
		level++
	}
	if level < 32 {
		return level
	}
	return 32
}
func (this *Set) HasKey(key string) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if _, ok := this.dict[key]; ok {
		return true
	}
	return false
}
