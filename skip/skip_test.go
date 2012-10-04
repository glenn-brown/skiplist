package skip

import (
	"fmt"
	"math/rand"
	"testing"
)

// Type Int is an int with Less() support.
//
type Int int

func (a Int) Less(b Lesser) bool {
	return a < b.(Int)
}

const (
	I1, I2, I3    = Int(1), Int(2), Int(3)
	I10, I20, I30 = Int(10), Int(20), Int(30)
)

func ExampleSkipList() {
	s := List{}
	s.Insert(I30, 3).Insert(I10, 1).Insert(I20, 2)
	for _, v, ok := s.RemoveN(0); ok; _, v, ok = s.RemoveN(0) {
		fmt.Println(v)
	}
	// Output: 1
	// 2
	// 3
}

func ExampleLenIndex() {
	s := List{}
	s.Insert(I30, 30).Insert(I30, 30).Insert(I20, 20)
	for i := s.Len() - 1; i >= 0; i-- {
		fmt.Println(s.PeekN(i))
	}
	// Output: 30
	// 20
	// 10
}

func TestSkipList(t *testing.T) {
	// Make a shuffled array integers from 0 to 99.
	ii := [100]Int{}
	for i := 0; i < 100; i++ {
		ii[i] = Int(i)
	}
	for i := 0; i < 100; i++ {
		r := rand.Intn(100)
		ii[i], ii[r] = ii[r], ii[i]
	}
	// Insert those entries into a skip list.
	s := List{}
	for i := 0; i < 100; i++ {
		s.Insert(ii[i], ii[i])
	}
	// Verify they are in order.
	for i := 0; i < 100; i++ {
		if p, ok := s.Peek(ii[i]); !ok || p.(Int) != ii[i] {
			panic("List is not in order")
		}
	}

	if _, v, ok := s.RemoveN(50); !ok || v.(Int) != Int(50) {
		panic("RemoveN failed")
	}
	if s.Len() != 99 {
		panic("Wrong Len")
	}
	if v, ok := s.Remove(Int(25)); !ok || v.(Int) != Int(25) {
		panic("Remove failed.")
	}
	if v, ok := s.Remove(Int(24)); !ok || v.(Int) != Int(24) {
		panic("Remove failed.")
	}
	if v, ok := s.Remove(Int(27)); !ok || v.(Int) != Int(28) {
		panic("Remove failed.")
	}
	if s.Len() != 98 {
		panic("Wrong Len")
	}
}

func ExampleString() {
	skip := List{}
	skip.Insert(I1, 10)
	skip.Insert(I2, 20)
	skip.Insert(I3, 30)
	fmt.Println(skip)
	// Output: {1:10 2:20 3:30}
}
