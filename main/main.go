package main

import (
	"fmt"
	"runtime"
	"sort"
	"time"
)

const M = 109 + 7
const (
	OffsetX1 = 0
	OffsetY1 = 1
	OffsetX2 = 2
	OffsetY2 = 3
)

type Store struct {
	vals     [][]int
	stack    []int
	sid, vid int
	v        uint64
	currentX uint64
}

func NewStore(vals [][]int) Store {
	sort.Slice(vals, func(i, j int) bool {
		return vals[i][OffsetX1] < vals[j][OffsetX1]
	})
	return Store{
		vals:  vals,
		stack: []int{},
		sid:   -1,
		vid:   0,
	}
}

func (s *Store) Spush(id, offset int) {
	f := s.Search(id, offset)
	s.stack = append(s.stack, id)
	copy(s.stack[f+1:], s.stack[f:len(s.stack)-1])
	s.sid = s.stack[0]
}

func (s *Store) Search(id, offset int) int {
	var l, r, i, v int = 0, len(s.stack) - 1, 0, s.vals[id][offset]
	for {
		if r-l <= 1 {
			return r
		}
		i = (l + r) / 2
		if s.vals[i][offset] < v {
			l = i
		} else {
			r = i
		}
	}
}

func (s *Store) Cal() uint64 {
	var sum uint64
	var v int
	for {
		if s.sid == -1 && s.vid == -1 {
			break
		}
		if s.sid == -1 {
			sum += s.add()
			continue
		}
		if s.vid == -1 {
			sum += s.cut()
			continue
		}
		v = s.vals[s.sid][OffsetX2] - s.vals[s.vid][OffsetX1]
		if v == 0 {
			sum += s.cutAndAdd()
			continue
		}
		if v > 0 {
			sum += s.add()
			continue
		}
		if v < 0 {
			sum += s.cut()
			continue
		}
	}
	return sum
}

// push vid值到store
func (s *Store) add() uint64 {
	var v uint64
	addX1 := s.vals[s.vid][OffsetX1]
	v = (uint64(addX1) - s.currentX) * s.v
	for i := s.vid; i < len(s.vals); i++ {
		if s.vals[i][OffsetX1] == addX1 {
			s.Spush(i, OffsetX1)
			s.vid = i
			continue
		}
		break
	}
	s.vid++
	if s.vid >= len(s.vals) {
		s.vid = -1
	}
	// cal v

	s.currentX = uint64(addX1)
	return v
}

// 从store->sid取出矩形
func (s *Store) cut() uint64 {
	var v uint64
	var cutX2 = s.vals[s.sid][OffsetX2]
	var i int
	for i = 0; i < len(s.stack); i++ {
		if s.vals[s.stack[i]][OffsetX2] != cutX2 {
			break
		}
	}
	v = s.v * (uint64(cutX2) - s.currentX)
	// cal v
	s.stack = s.stack[i:]
	if len(s.stack) == 0 {
		s.sid = -1
	}
	s.currentX = uint64(cutX2)
	return v
}

// 取出矩形，同时添加
func (s *Store) cutAndAdd() uint64 {
	var v uint64
	addX1 := s.vals[s.vid][OffsetX1]
	v = (uint64(addX1) - s.currentX) * s.v

	for i := s.vid; i < len(s.vals); i++ {
		if s.vals[i][OffsetX1] == addX1 {
			s.Spush(i, OffsetX1)
			s.vid = i
			continue
		}
		break
	}
	s.vid++
	if s.vid >= len(s.vals) {
		s.vid = -1
	}
	s.currentX = uint64(addX1)

	var cutX2 = s.vals[s.sid][OffsetX2]
	if addX1 != cutX2 {
		panic("addx1 and cutx2")
	}
	var i int
	for i = 0; i < len(s.stack); i++ {
		if s.vals[s.stack[i]][OffsetX2] != cutX2 {
			break
		}
	}
	// cal v

	s.stack = s.stack[i:]
	if len(s.stack) == 0 {
		s.sid = -1
	}
	s.currentX = uint64(cutX2)

	return v
}

func rectangleArea(rectangles [][]int) int {
	s := NewStore(rectangles)
	return int(s.Cal() % M)
}

func main() {
	m := make(map[uint64][128]byte)
	var i uint64 = 1_000_000
	var j uint64 = 0
	var cur uint64
	tk := time.After(time.Second * 10)
BK:
	for ; ; j++ {
		select {
		case <-tk:
			break BK
		default:
			for cur = 0; cur < i; cur++ {
				m[cur+i*j] = [128]byte{2}
			}
			for cur = 0; cur < i; cur++ {
				delete(m, cur+i*j)
			}
		}
	}
	for i1 := 0; i1 < 10; i1++ {
		j++
		m[j] = [128]byte{2}
		fmt.Println(len(m))
		runtime.GC()
		time.Sleep(time.Second)
	}
	m = make(map[uint64][128]byte)
	for {
		j++
		m[j] = [128]byte{2}
		fmt.Println(len(m))
		runtime.GC()
		time.Sleep(time.Second)
	}
}
