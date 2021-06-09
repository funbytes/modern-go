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

func (s *Set) createNode(key string, score float64) *node {
	n := new(node)
	n.key = key
	n.score = score
	return n
}

// 根据排名获取node 按升序获得 rank 从1开始
func (s *Set) GetElementByRankASC(rank int64) Node {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var n Node
	x := s.skipList.header
	var traversed int64 = 0
	for i := s.skipList.level - 1; i >= 0; i-- {
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

// 根据排名获取node 按降序获得 rank 从1开始
func (s *Set) GetElementByRankDESC(rank int64) Node {
	rank = s.GetLenth() - rank + 1
	return s.GetElementByRankASC(rank)
}
func (s *Set) GetScore(key string) (float64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if _, ok := s.dict[key]; !ok {
		return 0, errors.New("not find key " + key)
	}
	score := *s.dict[key]
	return score, nil
}

// 按升序排名 从1开始
func (s *Set) GetRank(key string) (int64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if _, ok := s.dict[key]; !ok {
		return 0, errors.New("not find key " + key)
	}
	var x *node
	var rank int64 = 0
	score := *s.dict[key]
	x = s.skipList.header
	for i := s.skipList.level - 1; i >= 0; i-- {
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

// 按降序排名
func (s *Set) GetRankDESC(key string) (int64, error) {
	rank, err := s.GetRank(key)
	if err != nil {
		return 0, err
	}
	return s.GetLenth() - rank + 1, nil
}

func (s *Set) Del(key string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.dict[key]; !ok {
		return errors.New("not find key " + key)
	}
	score := *s.dict[key]
	var update [DefaultMaxLevel]*node
	var x *node
	x = s.skipList.header
	var i int
	for i = s.skipList.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.key < key)) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	x = x.level[0].forward
	if x != nil && score == x.score && key == x.key {
		for i = 0; i < s.skipList.level; i++ {
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
			s.skipList.tail = x.backward
		}
		for s.skipList.level > 1 && s.skipList.header.level[s.skipList.level-1].forward == nil {
			s.skipList.level--
		}
		s.skipList.length--
	}
	delete(s.dict, key)
	return nil
}

func (s *Set) Set(key string, score float64) *node {
	s.Del(key)
	node := s.insert(key, score)
	return node
}

// 获取top n个数 按升序获取
func (s *Set) GetTopN(n int64) []Node {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var nodes []Node
	var q int64 = 0
	for x := s.skipList.header.level[0].forward; x != nil && q < n; x = x.level[0].forward {
		var node Node
		node.Key = x.key
		node.Score = x.score
		nodes = append(nodes, node)
		q++
	}
	return nodes
}

// 获取top n个数 按降序获取
func (s *Set) GetTopNDESC(n int64) []Node {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var nodes []Node
	var q int64 = 0
	for x := s.skipList.tail; x != nil && q < n; x = x.backward {
		var node Node
		node.Key = x.key
		node.Score = x.score
		nodes = append(nodes, node)
		q++
	}
	return nodes
}

func (s *Set) insert(key string, score float64) *node {
	s.lock.Lock()
	defer s.lock.Unlock()
	var update [DefaultMaxLevel]*node
	var x *node
	var rank [DefaultMaxLevel]int64
	var i, level int

	x = s.skipList.header

	for i = s.skipList.level - 1; i >= 0; i-- {
		rank[i] = 0
		if i != s.skipList.level-1 {
			rank[i] = rank[i+1]
		}
		for x.level[i].forward != nil && (x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.key < key)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}
	level = s.randomLevel()
	if level > s.skipList.level {
		for i = s.skipList.level; i < level; i++ {
			rank[i] = 0
			update[i] = s.skipList.header
			update[i].level[i].span = s.skipList.length
		}
		s.skipList.level = level
	}
	x = s.createNode(key, score)
	for i = 0; i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}
	for i = level; i < s.skipList.level; i++ {
		update[i].level[i].span++
	}
	if update[0] != s.skipList.header {
		x.backward = update[0]
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		s.skipList.tail = x
	}
	s.skipList.length++
	s.dict[key] = &x.score
	return x
}

func (s *Set) GetLenth() int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.skipList.length
}

func (s *Set) GetLevel() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.skipList.level
}

func (s *Set) randomLevel() int {
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
func (s *Set) HasKey(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if _, ok := s.dict[key]; ok {
		return true
	}
	return false
}
