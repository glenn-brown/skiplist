package skiplist

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

//
// Utility functions
//

func less(a, b interface{}) bool {
	return a.(int) < b.(int)
}

func shuffleRange(min, max int) []int {
	a := make ([]int, max - min + 1)
	for i := range(a) {
		a[i] = min+i
	}
	for i := range(a) {
		other := rand.Intn(max-min+1)
		a[i], a[other] = a[other], a[i]
	}
	return a
}

func skiplist(min, max int) *Skiplist {
	s := New(less, nil)
	for _, v := range shuffleRange(min,max) {
		s.Insert (v, 2*v)
	}
	return s
}

//
// Benchmarks, examples, and Tests
//

func TestSkiplist(t *testing.T) {
	s := skiplist(1, 20)
	i := 1
	for e := s.Front(); e != nil; e = e.Next() {
		if e.Key().(int) != i || e.Value.(int) != 2*i {
			t.Fail()
		}
		i++
	}
}

func TestElement_Key(t *testing.T) {
	e := skiplist(1,3).Front()
	for i := 1; i<=3; i++ {
		if e == nil || e.Key().(int) != i {
			t.Fail()
		}
		e = e.Next()
	}
}

func ExampleElement_Next() {
	s := New(less, nil).Insert(0, 0).Insert(1, 2).Insert(2, 4).Insert(3, 6)
	for e := s.Front(); e != nil; e = e.Next() {
		fmt.Print(e, " ")
	}
	// Output: 0:0 1:2 2:4 3:6
}

func TestElement_String(t *testing.T) {
	if fmt.Sprint(skiplist(1,2).Front()) != "1:2" {
		t.Fail()
	}
}

func TestNew(t *testing.T) {
	// Verify the injected random number generator is used.
	s := New(less, nil)
	s1 := New(less, rand.New(rand.NewSource(1)))
	s42 := New(less, rand.New(rand.NewSource(42)))
	for i:=0; i<32; i++ {
		s.Insert(i,i)
		s1.Insert(i,i)
		s42.Insert(i,i)
	}
	v := s.Visualization()
	v1 := s1.Visualization()
	v42 := s42.Visualization()
	if v == v1 {
		t.Error("Seed did not change behaviour")
	} else if v != v42 {
		t.Error("Default seed is not 42.")
	}
}

func TestSkiplist_Front(t *testing.T) {
	s := skiplist (1,3)
	if s.Front().Key().(int) != 1 {
		t.Fail()
	}
}

func TestSkiplist_Insert(t *testing.T) {
	if skiplist(1, 10).String() != "{1:2 2:4 3:6 4:8 5:10 6:12 7:14 8:16 9:18 10:20}" {
		t.Fail()
	}
}

func BenchmarkSkiplist_Insert(b *testing.B) {
	s := New(less, nil)
	for i:=0; i<b.N; i++ {
		s.Insert(i,i)
	}
}

func TestSkiplist_Remove(t *testing.T) {
	s := skiplist(0,10)
	if s.Remove(-1) != nil || s.Remove(11) != nil {
		t.Error("Removing nonexistant key should fail.")
	}
	for i:= range shuffleRange(0,10) {
		e := s.Remove(i)
		if e == nil {
			t.Error("nil")
		}
		if e.Key().(int) != i {
			t.Error("bad key")
		}
		if e.Value.(int) != 2*i {
			t.Error("bad value")
		}
	}
	if s.Len() != 0 {
		t.Error("nonzero len")
	}
}

func TestSkiplist_RemoveN(t *testing.T) {
	s := skiplist(0,10)
	keys := shuffleRange(0,10)
	cnt := 11
	for _,key := range(keys) {
		found, pos := s.Find(key)
		t.Log(key, found, pos)
		t.Log("\n" + s.Visualization())
		e := s.RemoveN(pos)
		if e == nil {
			t.Error("nil returned")
		} else if found != e {
			t.Error("Wrong removed")
		} else if e.Key().(int) != key {
			t.Error("bad Key()")
		} else if e.Value.(int) != 2*key {
			t.Error("bad Value")
		}
		cnt--
		l := s.Len()
		if l != cnt {
			t.Error ("bad Len()=", l, "!=", cnt)
		}
	}
}

func TestSkiplist_Find(t *testing.T) {
	s := skiplist(0, 9)
	for i := s.Len()-1; i>=0; i-- {
		e, pos := s.Find(i)
		if e == nil {
			t.Error("nil")
		} else if e != s.FindN(pos) {
			t.Error("bad pos")
		} else if e.Key().(int) != i {
			t.Error("bad Key")
		} else if e.Value.(int) != 2*i {
			t.Error ("bad Value")
		}
	}
}
	
func TestSkiplist_Len(t *testing.T) {
	s := skiplist(0, 4)
	if s.Len() != 5 {
		t.Fail()
	}
}

func TestSkiplist_FindN(t *testing.T) {
	s := skiplist(0, 9)
	for i := s.Len()-1; i>=0; i-- {
		e := s.FindN(i)
		if e == nil {
			t.Error("nil")
		} else if e.Key().(int) != i {
			t.Error("bad Key")
		} else if e.Value.(int) != 2*i {
			t.Error ("bad Value")
		}
	}
}

func ExampleSkiplist_String() {
	skip := New(less, nil).Insert(1, 10).Insert(2, 20).Insert(3, 30)
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
