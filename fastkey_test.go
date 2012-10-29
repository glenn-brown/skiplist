// Copyright 2012 The Skiplist Authors

package skiplist

import "fmt"

// For any old type:
type FastType struct{ a, b int }

// Implement the SlowKey interface:
func (a *FastType) Less(b interface{}) bool {
	// For example, sort by the sum of the elements in the struct:
	mb := b.(*FastType)
	return (a.a + a.b) < (mb.a + mb.b)
}
func (*FastType) Score(i interface{}) float64 {
	// Score(i) increase monotonically with increasing key value.
	m := i.(*FastType)
	return float64(m.a + m.b)
}

// Any type implementing the FastKey interface can be used as a key.
//
func ExampleFastKey() {
	keys := []FastType{{1, 2}, {5, 6}, {3, 4}}
	s := New().Insert(&keys[0], 1).Insert(&keys[1], 2).Insert(&keys[2], 3)
	fmt.Print(s)
	// Output: {&{1 2}:1 &{3 4}:3 &{5 6}:2}
}
