// Copyright (c) 2012, Glenn Brown.  All rights reserved.  See LICENSE.

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
// Score implements the FastKey interface.
func (m *FastType) Score() float64 {
	return float64(m.a + m.b)
}

// Any type implementing the FastKey interface can be used as a key.
//
func ExampleFastKey() {
	keys := []FastType{{1, 2}, {5, 6}, {3, 4}}
	s := New().Insert(FastKey(&keys[0]), 1).Insert(FastKey(&keys[1]), 2).Insert(FastKey(&keys[2]), 3)
	fmt.Print(s)
	// Output: {&{1 2}:1 &{3 4}:3 &{5 6}:2}
}
