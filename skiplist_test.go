// Copyright 2012 The Skiplist Authors

package skiplist

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

////////////////////////////////////////////////////////////////
// Tests
////////////////////////////////////////////////////////////////

func TestSkiplist(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	e := skiplist(1, 3).Front()
	for i := 1; i <= 3; i++ {
		if e == nil || e.Key().(int) != i {
			t.Fail()
		}
		e = e.Next()
	}
}

func TestElement_String(t *testing.T) {
	t.Parallel()
	if fmt.Sprint(skiplist(1, 2).Front()) != "1:2" {
		t.Fail()
	}
}

func TestNew(t *testing.T) {
	t.Parallel()
	// Verify the injected random number generator is used.
	s := New(nil)
	s1 := New(rand.New(rand.NewSource(1)))
	s42 := New(rand.New(rand.NewSource(42)))
	for i := 0; i < 32; i++ {
		s.Insert(i, i)
		s1.Insert(i, i)
		s42.Insert(i, i)
	}
	v := s.Visualization()
	v1 := s1.Visualization()
	v42 := s42.Visualization()
	if v == v1 {
		t.Error("Seed did not change behaviour")
	} else if v != v42 {
		// t.Error("Default seed is not 42.")
	}
}

func TestSkiplist_Front(t *testing.T) {
	t.Parallel()
	s := skiplist(1, 3)
	if s.Front().Key().(int) != 1 {
		t.Fail()
	}
}

func TestSkiplist_Insert(t *testing.T) {
	t.Parallel()
	if skiplist(1, 10).String() != "{1:2 2:4 3:6 4:8 5:10 6:12 7:14 8:16 9:18 10:20}" {
		t.Fail()
	}
}

func TestSkiplist_RemoveHead(t *testing.T) {
	t.Parallel()
	s := skiplist(0, 10)
	if s.Remove(-1) != nil || s.Remove(11) != nil {
		t.Error("Removing nonexistant key should fail.")
	}
	for i := range shuffleRange(0, 10) {
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
	t.Parallel()
	s := skiplist(0, 10)
	keys := shuffleRange(0, 10)
	cnt := 11
	for _, key := range keys {
		found, pos := s.ElementPos(key)
		t.Logf("Removing key=%v at pos=%v", key, pos)
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
			t.Error("bad Len()=", l, "!=", cnt)
		}
	}
}

func TestSkiplist_Element_forward(t *testing.T) {
	t.Parallel()
	s := skiplist(0, 9)
	for i := s.Len() - 1; i >= 0; i-- {
		e, pos := s.ElementPos(i)
		if e == nil {
			t.Error("nil")
		} else if e != s.ElementN(pos) {
			t.Error("bad pos")
		} else if e.Key().(int) != i {
			t.Error("bad Key")
		} else if e.Value.(int) != 2*i {
			t.Error("bad Value")
		}
	}
}

func TestSkiplist_Len(t *testing.T) {
	t.Parallel()
	s := skiplist(0, 4)
	if s.Len() != 5 {
		t.Fail()
	}
}

func TestSkiplist_ElementN(t *testing.T) {
	t.Parallel()
	s := skiplist(0, 9)
	for i := s.Len() - 1; i >= 0; i-- {
		e := s.ElementN(i)
		if e == nil {
			t.Error("nil")
		} else if e.Key().(int) != i {
			t.Error("bad Key")
		} else if e.Value.(int) != 2*i {
			t.Error("bad Value")
		}
	}
}

func TestBuiltins(t *testing.T) {
	t.Parallel()

	// Create high and low variables for each ordered builtin type.

	f32a, f32b := float32(1.0), float32(0.0)
	f64a, f64b := float64(1.0), float64(0.0)
	i16a, i16b := int16(1), int16(0)
	i32a, i32b := int32(1), int32(0)
	i64a, i64b := int64(1), int64(0)
	i8_a, i8_b := int8(1), int8(0)
	i__a, i__b := int(1), int(0)
	sl_a, sl_b := []byte{1}, []byte{0}
	stra, strb := "1", "0"
	u16a, u16b := uint16(1), uint16(0)
	u32a, u32b := uint32(1), uint32(0)
	u64a, u64b := uint64(1), uint64(0)
	u8_a, u8_b := uint8(1), uint8(0)
	u__a, u__b := uint(1), uint(0)
	up_a, up_b := uintptr(1), uintptr(0)

	// Insert pairs in a map and verify the large is in position 1.

	if New(nil).Set(f32a, 1).Set(f32b, 2).Pos(f32a) != 1 {
		t.Error("float32")
	}
	if New(nil).Set(f64a, 1).Set(f64b, 2).Pos(f64a) != 1 {
		t.Error("float64")
	}
	if New(nil).Set(i16a, 1).Set(i16b, 2).Pos(i16a) != 1 {
		t.Error("int16")
	}
	if New(nil).Set(i32a, 1).Set(i32b, 2).Pos(i32a) != 1 {
		t.Error("int32")
	}
	if New(nil).Set(i64a, 1).Set(i64b, 2).Pos(i64a) != 1 {
		t.Error("int64")
	}
	if New(nil).Set(i8_a, 1).Set(i8_b, 2).Pos(i8_a) != 1 {
		t.Error("int8")
	}
	if New(nil).Set(i__a, 1).Set(i__b, 2).Pos(i__a) != 1 {
		t.Error("int")
	}
	if New(nil).Set(sl_a, 1).Set(sl_b, 2).Pos(sl_a) != 1 {
		t.Error("[]byte")
	}
	if New(nil).Set(stra, 1).Set(strb, 2).Pos(stra) != 1 {
		t.Error("string")
	}
	if New(nil).Set(u16a, 1).Set(u16b, 2).Pos(u16a) != 1 {
		t.Error("uint16")
	}
	if New(nil).Set(u32a, 1).Set(u32b, 2).Pos(u32a) != 1 {
		t.Error("uint32")
	}
	if New(nil).Set(u64a, 1).Set(u64b, 2).Pos(u64a) != 1 {
		t.Error("uint64")
	}
	if New(nil).Set(u8_a, 1).Set(u8_b, 2).Pos(u8_a) != 1 {
		t.Error("uint8")
	}
	if New(nil).Set(u__a, 1).Set(u__b, 2).Pos(u__a) != 1 {
		t.Error("uint")
	}
	if New(nil).Set(up_a, 1).Set(up_b, 2).Pos(up_a) != 1 {
		t.Error("uintptr")
	}

	// Insert pairs in a map and verify the large is in position 1.

	if NewDescending(nil).Set(f32a, 1).Set(f32b, 2).Pos(f32b) != 1 {
		t.Error("float32")
	}
	if NewDescending(nil).Set(f64a, 1).Set(f64b, 2).Pos(f64b) != 1 {
		t.Error("float64")
	}
	if NewDescending(nil).Set(i16a, 1).Set(i16b, 2).Pos(i16b) != 1 {
		t.Error("int16")
	}
	if NewDescending(nil).Set(i32a, 1).Set(i32b, 2).Pos(i32b) != 1 {
		t.Error("int32")
	}
	if NewDescending(nil).Set(i64a, 1).Set(i64b, 2).Pos(i64b) != 1 {
		t.Error("int64")
	}
	if NewDescending(nil).Set(i8_a, 1).Set(i8_b, 2).Pos(i8_b) != 1 {
		t.Error("int8")
	}
	if NewDescending(nil).Set(i__a, 1).Set(i__b, 2).Pos(i__b) != 1 {
		t.Error("int")
	}
	if NewDescending(nil).Set(sl_a, 1).Set(sl_b, 2).Pos(sl_b) != 1 {
		t.Error("[]byte")
	}
	if NewDescending(nil).Set(stra, 1).Set(strb, 2).Pos(strb) != 1 {
		t.Error("string")
	}
	if NewDescending(nil).Set(u16a, 1).Set(u16b, 2).Pos(u16b) != 1 {
		t.Error("uint16")
	}
	if NewDescending(nil).Set(u32a, 1).Set(u32b, 2).Pos(u32b) != 1 {
		t.Error("uint32")
	}
	if NewDescending(nil).Set(u64a, 1).Set(u64b, 2).Pos(u64b) != 1 {
		t.Error("uint64")
	}
	if NewDescending(nil).Set(u8_a, 1).Set(u8_b, 2).Pos(u8_b) != 1 {
		t.Error("uint8")
	}
	if NewDescending(nil).Set(u__a, 1).Set(u__b, 2).Pos(u__b) != 1 {
		t.Error("uint")
	}
	if NewDescending(nil).Set(up_a, 1).Set(up_b, 2).Pos(up_b) != 1 {
		t.Error("uintptr")
	}
}

////////////////////////////////////////////////////////////////
// Examples
////////////////////////////////////////////////////////////////

func Example() {
	// Create a skiplist and add some entries:
	s := New(nil).Set("one", "un").Set("two", nil).Set("three", "trois")

	// Retrieve a mapping:
	fmt.Println(s.Get("two"))

	// Replace a mapping:
	s.Set("two", "deux")

	// Print the skiplist:
	fmt.Println(s)

	// Add more than one value for a key, even of different value-type:
	s.Insert("three", 3)

	// Retrieve all values for the key:
	fmt.Println(s.GetAll("three"))

	// Or just the youngest:
	fmt.Println(s.Get("three"))

	// Iterate over all values in the map:
	for e := s.Front(); nil != e; e = e.Next() {
		fmt.Print(e.Key(), "->", e.Value, ",")
	}
	fmt.Println()

	// Pop the first entry:
	s.RemoveN(0)

	// Pop the last entry:
	s.RemoveN(s.Len() - 1)
	fmt.Println(s)

	// Output:
	// <nil>
	// {one:un three:trois two:deux}
	// [3 trois]
	// 3
	// one->un,three->3,three->trois,two->deux,
	// {three:3 three:trois}
}

// This example demonstrates iteration over all list elements.
func ExampleElement_Next() {
	s := New(nil).Insert(0, 0).Insert(1, 2).Insert(2, 4).Insert(3, 6)
	for e := s.Front(); e != nil; e = e.Next() {
		fmt.Print(e, " ")
	}
	// Output: 0:0 1:2 2:4 3:6
}

func ExampleSkiplist_String() {
	skip := New(nil).Insert(1, 10).Insert(2, 20).Insert(3, 30)
	fmt.Println(skip)
	// Output: {1:10 2:20 3:30}
}

func TestVisualization(t *testing.T) {
	s := New(nil)
	for i := 0; i < 23; i++ {
		s.Insert(i, i)
	}
	v := s.Visualization()
	expected := "" +
		"L4 |------------------------------------------------------------------->|->/\n" +
		"L3 |------------------------------------------------------------------->|->/\n" +
		"L2 |---------->|---->|------->|------------------->|------------->|---->|->/\n" +
		"L1 |---->|---->|---->|------->|------------------->|------->|---->|---->|->/\n" +
		"L0 |->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->/\n" +
		"      0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  1  1  1  1  1  1  1\n" +
		"      0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f  0  1  2  3  4  5  6"
	if v != expected {
		t.Error(v, "\n!=\n", expected)
	}
}

////////////////////////////////////////////////////////////////
// Benchmarks
////////////////////////////////////////////////////////////////

func BenchmarkSkiplist_Insert_forward(b *testing.B) {
	b.StopTimer()
	s := New(nil)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Insert(i, i)
	}
}

func BenchmarkSkiplist_Insert_reverse(b *testing.B) {
	b.StopTimer()
	s := New(nil)
	b.StartTimer()
	for i := b.N - 1; i >= 0; i-- {
		s.Insert(i, i)
	}
}

func BenchmarkSkiplist_Insert_shuffle(b *testing.B) {
	b.StopTimer()
	a := shuffleRange(0, b.N-1)
	s := New(nil)
	b.StartTimer()
	for i, key := range a {
		s.Insert(key, i)
	}
}

func BenchmarkSkiplist_Element_forward(b *testing.B) {
	b.StopTimer()
	s := New(nil)
	for i := b.N - 1; i >= 0; i-- {
		s.Insert(i, i)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Element(i)
	}
}

func BenchmarkSkiplist_Element_reverse(b *testing.B) {
	b.StopTimer()
	s := New(nil)
	for i := 0; i < b.N; i++ {
		s.Insert(i, i)
	}
	b.StartTimer()
	for i := b.N - 1; i >= 0; i-- {
		s.Element(i)
	}
}

func BenchmarkSkiplist_Element_shuffle(b *testing.B) {
	b.StopTimer()
	a := shuffleRange(0, b.N-1)
	s := skiplist(0, b.N-1)
	b.StartTimer()
	for _, key := range a {
		s.Element(key)
	}
}

func BenchmarkSkiplist_ElementN_forward(b *testing.B) {
	b.StopTimer()
	s := New(nil)
	for i := b.N - 1; i >= 0; i-- {
		s.Insert(i, i)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.ElementN(i)
	}
}

func BenchmarkSkiplist_ElementN_reverse(b *testing.B) {
	b.StopTimer()
	s := New(nil)
	for i := 0; i < b.N; i++ {
		s.Insert(i, i)
	}
	b.StartTimer()
	for i := b.N - 1; i >= 0; i-- {
		s.ElementN(i)
	}
}

func BenchmarkSkiplist_ElementN_shuffle(b *testing.B) {
	b.StopTimer()
	a := shuffleRange(0, b.N-1)
	s := skiplist(0, b.N-1)
	b.StartTimer()
	for _, key := range a {
		s.ElementN(key)
	}
}

func BenchmarkSkiplist_Remove_forward(b *testing.B) {
	b.StopTimer()
	s := New(nil)
	for i := b.N - 1; i >= 0; i-- {
		s.Insert(i, i)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Remove(i)
	}
}

func BenchmarkSkiplist_Remove_reverse(b *testing.B) {
	b.StopTimer()
	s := New(nil)
	for i := 0; i < b.N; i++ {
		s.Insert(i, i)
	}
	b.StartTimer()
	for i := b.N - 1; i >= 0; i-- {
		s.Remove(i)
	}
}

func BenchmarkSkiplist_Remove_shuffle(b *testing.B) {
	b.StopTimer()
	a := shuffleRange(0, b.N-1)
	s := skiplist(0, b.N-1)
	b.StartTimer()
	for _, key := range a {
		s.Remove(key)
	}
}

func BenchmarkSkiplist_RemoveN_head(b *testing.B) {
	b.StopTimer()
	s := New(nil)
	for i := b.N - 1; i >= 0; i-- {
		s.Insert(i, i)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.RemoveN(0)
	}
}

func BenchmarkSkiplist_RemoveN_tail(b *testing.B) {
	b.StopTimer()
	s := New(nil)
	for i := 0; i < b.N; i++ {
		s.Insert(i, i)
	}
	b.StartTimer()
	for i := b.N - 1; i >= 0; i-- {
		s.RemoveN(i)
	}
}

func BenchmarkSkiplist_RemoveN_mid(b *testing.B) {
	b.StopTimer()
	s := skiplist(0, b.N-1)
	b.StartTimer()
	for i := b.N - 1; i >= 0; i-- {
		s.RemoveN(i / 2)
	}
}

////////////////////////////////////////////////////////////////
// Utility functions
////////////////////////////////////////////////////////////////

// Create a shuffled slice of the integers in [min,max].
//
func shuffleRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	for i := range a {
		other := rand.Intn(max - min + 1)
		a[i], a[other] = a[other], a[i]
	}
	return a
}

// Create a Skiplist with each key in [min,max].
//
func skiplist(min, max int) *Skiplist {
	s := New(nil)
	for _, v := range shuffleRange(min, max) {
		s.Insert(v, 2*v)
	}
	return s
}

// Create an arrow string like "|-->" that is cnt runes long.
//
func arrow(cnt int) (s string) {
	cnt *= 3
	switch {
	case cnt > 1:
		return "|" + strings.Repeat("-", cnt-2) + ">"
	case cnt == 1:
		return ">"
	}
	return "X"
}

// Create a visualization string like this:
//   Output:
//   L4 |---------------------------------------------------------------------->/
//   L3 |------------------------------------------->|------------------------->/
//   L2 |---------->|---------->|---------->|------->|---------------->|---->|->/
//   L1 |---------->|---------->|---------->|->|---->|->|->|->|------->|->|->|->/
//   L0 |->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->/
//         0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  1  1  1  1  1  1  1  
//         0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f  0  1  2  3  4  5  6
//
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
	s += "    "
	for n := l.links[0].to; n != nil; n = n.links[0].to {
		s += fmt.Sprintf("  %x", n.key.(int)>>4&0xf)
	}
	s += "\n    "
	for n := l.links[0].to; n != nil; n = n.links[0].to {
		s += fmt.Sprintf("  %x", n.key.(int)&0xf)
	}
	return string(s)
}
