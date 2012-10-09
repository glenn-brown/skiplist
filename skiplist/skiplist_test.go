package skiplist

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func less(a, b interface{}) bool {
	return a.(int) < b.(int)
}

func ExampleSkiplist() {
	s := New(less, nil)
	s.Insert(30, 3).Insert(10, 1).Insert(20, 2)
	for e := s.RemoveN(0); e != nil; e = s.RemoveN(0) {
		fmt.Println(e.Value)
	}
	// Output:
	// 1
	// 2
	// 3
}

func ExampleLenIndex() {
	s := New(less, nil)
	s.Insert(30, 30).Insert(20, 20).Insert(10, 10)
	for i := s.Len() - 1; i >= 0; i-- {
		fmt.Println(s.FindN(i))
	}
	// Output:
	// 30:30
	// 20:20
	// 10:10
}

func TestSkiplist(t *testing.T) {
	// Make a shuffled array integers from 0 to N-1
	ii := [64]int{}
	N := len(ii)
	for i := 0; i < N; i++ {
		ii[i] = i
	}
	for i := 0; i < N; i++ {
		r := rand.Intn(N)
		ii[i], ii[r] = ii[r], ii[i]
	}

	// Insert those entries into a skip list.
	s := New(less, nil)
	for i := 0; i < N; i++ {
		s.Insert(ii[i], ii[i])
	}
	/* fmt.Println(s.Visualization()) */
	if s.Len() != N {
		panic("Len")
	}

	// Find all entries by key.
	for i := 0; i < N; i++ {
		if e, _ := s.Find(ii[i]); e == nil || e.Value.(int) != ii[i] {
			panic("Find")
		}
	}

	// Find all entries by position.
	for i := 0; i < N; i++ {
		if e := s.FindN(i); e == nil || e.Value.(int) != i || e.Key().(int) != i {
			panic("FindN")
		}
	}

	// Verify they are in order.
	if e := s.RemoveN(50); e == nil || e.Value.(int) != 50 {
		panic("RemoveN failed")
	}
	if s.Len() != N-1 {
		panic("Wrong Len")
	}
	if e := s.Remove(25); e == nil || e.Value.(int) != 25 {
		panic("Remove(25) failed.")
	}
	if e := s.Remove(24); e == nil || e.Value.(int) != 24 {
		panic("Remove(24) failed.")
	}
	if e := s.Remove(27); e == nil || e.Value.(int) != 27 {
		panic("Remove(27) failed.")
	}
	if s.Len() != N-4 {
		panic("Wrong Len")
	}
}

func ExampleString() {
	skip := New(less, nil)
	skip.Insert(1, 10)
	skip.Insert(2, 20)
	skip.Insert(3, 30)
	fmt.Println(skip)
	// Output: {1:10 2:20 3:30}
}

func ExampleVisualization() {
	s := New(less, nil)
	for i := 0; i < 64; i++ {
		s.Insert(i, i)
	}
	fmt.Println(s.Visualization())
	// Output:
	// L6 ---------------------------------------------------------------->/
	// L5 ---------------------------------------------------------------->/
	// L4 ------------------------------------------------>----->--------->/
	// L3 -------->--------->----->----------------------->----->--------->/
	// L2 ----->-->--->----->----->----->->------>-------->----->--------->/
	// L1 ->>-->>->>->>--->>>->>->>----->->-->>->>-->>>->>>->>-->>---->-->>/
	// L0 >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>/
	//    0000000000000000111111111111111122222222222222223333333333333333
	//    0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
}

func arrow(cnt int) (s string) {
	switch {
	case cnt > 1:
		return strings.Repeat("-", cnt-1) + ">"
	case cnt == 1:
		return ">"
	}
	return "X"
}

func (l *Skiplist) Visualization() (s string) {
	for level := len(l.links) - 1; level >= 0; level-- {
		s += fmt.Sprintf("L%d ", level)
		w := l.links[level].width
		s += arrow(w)
		for n := l.links[level].to; n != nil; n = n.links[level].to {
			w = n.links[level].width
			s += arrow(w)
		}
		s += "/\n"
	}
	s += "   "
	for n := l.links[0].to; n != nil; n = n.links[0].to {
		s += fmt.Sprintf("%x", n.key.(int)>>4&0xf)
	}
	s += "\n   "
	for n := l.links[0].to; n != nil; n = n.links[0].to {
		s += fmt.Sprintf("%x", n.key.(int)&0xf)
	}
	return string(s)
}
