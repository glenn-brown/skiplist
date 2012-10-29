// Copyright 2012 The Skiplist Authors

package skiplist

import "fmt"

// For any old type:
type MyType struct{ a, b int }

// Implement the SlowKey interface:
func (a *MyType) Less(b interface{}) bool {
	// For example, sort by the sum of the elements in the struct:
	mb := b.(*MyType)
	return (a.a + a.b) < (mb.a + mb.b)
}

// Any type implementing the SlowKey interface can be used as a key.
//
func ExampleSlowKey() {
	keys := []MyType{{1, 2}, {5, 6}, {3, 4}}
	s := New().Insert(&keys[0], 1).Insert(&keys[1], 2).Insert(&keys[2], 3)
	fmt.Print(s)
	// Output: {&{1 2}:1 &{3 4}:3 &{5 6}:2}
}
