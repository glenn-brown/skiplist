// Copyright (c) 2012, Glenn Brown.  All rights reserved.  See LICENSE.

// Package skiplist implements fast indexable ordered multimaps.
//
// This skip list has some features that make it special:
// It supports position-index addressing.
// It can act as a map or as a multimap.
// It automatically adjusts its depth.
// It mimics Go's container/list interface where possible.
// It automatically and efficiently supports int*, float*, uint*, string, and []byte keys.
// It supports externally defined key types via the FastKey and SlowKey interfaces.
//
// Get, Set, Insert, Remove*, Element*, and Pos operations all require
// O(log(N)) time or less, where N is the number of entries in the
// list.  GetAll() requires O(log(N)+V) time where V is the number
// of values returned. The skiplist requires O(N) space.
//
package skiplist

import (
	"bytes"
	"fmt"
	"github.com/glenn-brown/ordinal"
	"math/rand"
)

// A Skiplist is linked at multiple levels.  The bottom level (L0) is
// a sorted linked list of entries, and each link has a link at the
// next higher level added with probability P at insertion.  Since
// this is a position-addressable skip-list, each link has an
// associated 'width' specifying the number of nodes it skips, so
// nodes can also be referenced by position.
//
// For example, a skiplist containing values from 0 to 0x16 might be structured
// like this:
//   L4 |---------------------------------------------------------------------->/
//   L3 |------------------------------------------->|------------------------->/
//   L2 |---------->|---------->|---------->|------->|---------------->|---->|->/
//   L1 |---------->|---------->|---------->|->|---->|->|->|->|------->|->|->|->/
//   L0 |->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->/
//         0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  1  1  1  1  1  1  1  
//         0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f  0  1  2  3  4  5  6
// The skiplist is searched starting at the top level, going as far right as possible
// without passing the desired Element, dropping down one level, and repeating for 
// each level.	
//
type Skiplist struct {
	cnt   int
	less  func(a, b interface{}) bool
	links []link
	prev  []prev
	rng   *rand.Rand
	score func(a interface{}) float64
}
type link struct {
	to    *Element
	width int
}

// Element is an key/value pair inserted into the list.  Use
// element.Key() to access the protected key.
//
type Element struct {
	key   interface{} // private to protect order
	Value interface{}
	score float64
	links []link
}

// Key returns the key used to insert the value in the list element in O(1) time.
//
func (e *Element) Key() interface{} { return e.key }

// Next returns the next-higher-indexed list element or nil in O(1) time.
//
func (e *Element) Next() *Element { return e.links[0].to }

// String returns a Key:Value string representation of the element.
//
func (e *Element) String() string { return fmt.Sprintf("%v:%v", e.key, e.Value) }

// New returns a new skiplist in O(1) time.
// The list will be sorted from least to greatest key.
//
func New() *Skiplist {
	nu := &Skiplist{}

	// Seed a private random number generator for reproducibility.

	nu.rng = rand.New(rand.NewSource(42))

	// Arrange to set nu.less and nu.score the first time either is called.
	// We can't do it here because we can't infer the key type until the first
	// key is inserted.

	nu.less = func(a, b interface{}) bool {
		nu.less, nu.score = ordinal.Fns(a)
		return nu.less(a, b)
	}
	nu.score = func(a interface{}) float64 {
		nu.less, nu.score = ordinal.Fns(a)
		return nu.score(a)
	}
	return nu
}

// NewDescending is like New, except keys are sorted from greatest to least.
//
func NewDescending() *Skiplist {
	nu := &Skiplist{}

	// Seed a private random number generator for reproducibility.

	nu.rng = rand.New(rand.NewSource(42))

	// Arrange to set nu.less and nu.score the first time either is called.
	// We can't do it here because we can't infer the key type until the first
	// key is inserted.

	nu.less = func(a, b interface{}) bool {
		nu.less, nu.score = ordinal.FnsReversed(a)
		return nu.less(a, b)
	}
	nu.score = func(a interface{}) float64 {
		nu.less, nu.score = ordinal.FnsReversed(a)
		return nu.score(a)
	}
	return nu
}

// Return the first list element in O(1) time.
//
func (l *Skiplist) Front() *Element {
	if len(l.links) == 0 {
		return nil
	}
	return l.links[0].to
}

// Insert a {key,value} pair in the skiplist, optionally replacing the youngest previous entry.
//
func (l *Skiplist) insert(key interface{}, value interface{}, replace bool) *Skiplist {
	l.grow()
	s := l.score(key)
	prev, pos := l.prevs(key, s)
	next := prev[0].link.to
	if replace && nil != next && s == next.score &&
		!l.less(key, next.key) && !l.less(next.key, key) {

		l.remove(prev, next)
	}
	nuLevels := l.randLevels(len(l.links))
	nu := &Element{key, value, s, make([]link, nuLevels)}
	for level := range prev {
		if level < nuLevels {
			if level == 0 {
				// At the bottom level, simply link in the new Element of width 1
				to := prev[level].link.to
				prev[level].link.to = nu
				nu.links[level].width = 1
				nu.links[level].to = to
				continue
			}
			// Link in the new element.
			end := prev[level].pos + prev[level].link.width + 1
			nu.links[level].to = prev[level].link.to
			nu.links[level].width = end - pos
			prev[level].link.to = nu
			prev[level].link.width = pos - prev[level].pos
			continue
		}
		// Higher levels just get a width adjustment.
		prev[level].link.width += 1
	}
	return l
}

// Insert a {key,value} pair into the skip list in O(log(N)) time.
//
func (l *Skiplist) Insert(key interface{}, value interface{}) *Skiplist {
	return l.insert(key, value, false)
}

// Get returns the value corresponding to key in the table in O(log(N)) time.
// If there is no corresponding value, nil is returned.
// If there are multiple corresponding values, the youngest is returned.
//
// If the list might contain an nil value, you may want to use GetOk instead.
//
func (l *Skiplist) Get(key interface{}) (value interface{}) {
	e, _ := l.ElementPos(key)
	if nil == e {
		return nil
	}
	return e.Value
}

// GetOk returns the value corresponding to key in the table in O(log(N)) time.
// The return value ok is true iff the key was present.
// If there is no corresponding value, nil and false are returned.
// If there are multiple corresponding values, the youngest is returned.
//
func (l *Skiplist) GetOk(key interface{}) (value interface{}, ok bool) {
	e, _ := l.ElementPos(key)
	if nil == e {
		return nil, false
	}
	return e.Value, true
}

// GetAll returns all values coresponding to key in the list, starting with the youngest.
// If no value corresponds, an empty slice is returned.
// O(log(N)+V) time is required, where M is the number of values returned.
//
func (l *Skiplist) GetAll(key interface{}) (values []interface{}) {
	s := l.score(key)
	prevs, _ := l.prevs(key, s)
	e := prevs[0].link.to
	for nil != e && e.score == s && !l.less(key, e.key) {
		values = append(values, e.Value)
		e = e.links[0].to
	}
	return values
}

// Insert a {key,value} pair into the skip list in O(log(N)) time, replacing the youngest entry
// for key, if any.
//
func (l *Skiplist) Set(key interface{}, value interface{}) *Skiplist {
	return l.insert(key, value, true)
}

// Function remove removes Element elem from a list.  Parameter prevs must be
// the precomputed predecessor list for the element.
//
func (l *Skiplist) remove(prev []prev, elem *Element) *Element {
	// At the bottom level, simply unlink the element.
	prev[0].link.to = elem.links[0].to
	// Unlink any higher linked levels.
	level := 1
	levels := len(l.links)
	for ; level < levels && prev[level].link.to == elem; level++ {
		prev[level].link.to = elem.links[level].to
		prev[level].link.width += elem.links[level].width - 1
	}
	// Adjust widths at higher levels
	for ; level < levels; level++ {
		prev[level].link.width -= 1
	}
	l.shrink()
	return elem
}

// Remove the youngest Element associate with Key, if any, in O(log(N)) time.
// Return the removed element or nil.
//
func (l *Skiplist) Remove(key interface{}) *Element {
	s := l.score(key)
	prevs, _ := l.prevs(key, s)
	// Verify there is a matching entry to remove.
	elem := l.prev[0].link.to
	if elem == nil || s != elem.score || s == elem.score && l.less(key, elem.key) {
		return nil
	}
	return l.remove(prevs, elem)
}

// Remove the specified element from the table, in O(log(N)) time.
// If the element is one of M multiple entries for the key, and additional O(M) time is required.
// This is useful for removing a specific element in a multimap, or removing elements during iteration.
//
func (l *Skiplist) RemoveElement(e *Element) *Element {

	// Find the first element in the multimap group.

	k := e.key
	s := l.score(k)
	prevs, pos := l.prevs(k, s)

	// Find the position of the matching entry within the multimap group.

	for match := prevs[0].link.to; nil != match && match != e; match = match.Next() {
		pos++
	}

	// Adjust prevs to be relative to the element, not relative to the start of the group.

	levels := len(prevs)
	for level := 0; level < levels; level++ {
		for l := prevs[level]; l.pos+l.link.width < pos; {
			prevs[level].pos = l.pos + l.link.width
			prevs[level].link = &l.link.to.links[level]
		}
	}

	// Remove the element.

	return l.remove(prevs, e)
}

// RemoveN removes any element at position pos in O(log(N)) time,
// returning it or nil.
//
func (l *Skiplist) RemoveN(index int) *Element {
	if index >= l.cnt {
		return nil
	}
	prevs := l.prevsN(index)
	elem := prevs[0].link.to
	return l.remove(prevs, elem)
}

// Element returns the youngest list element for key and its position,
// If there is no match, nil and -1 are returned.
//
// Consider using Get or GetAll instead if you only want Values.
//
func (l *Skiplist) ElementPos(key interface{}) (e *Element, pos int) {
	s := l.score(key)
	prev, pos := l.prevs(key, s)
	elem := prev[0].link.to
	if elem == nil || s < elem.score || s == elem.score && l.less(key, elem.key) {
		return nil, -1
	}
	return elem, pos
}

// Element returns the youngest list element for key,
// without modifying the list, in O(log(N)) time.
// If there is no match, nil is returned.
//
func (l *Skiplist) Element(key interface{}) (e *Element) {
	e, _ = l.ElementPos(key)
	return e
}

// ElementPos returns the position of the youngest list element for key,
// without modifying the list, in O(log(N)) time.
// If there is no match, -1 is returned.
//
// Consider using Get or GetAll instead if you only want Values.
//
func (l *Skiplist) Pos(key interface{}) (pos int) {
	_, pos = l.ElementPos(key)
	return pos
}

// Len returns the number of elements in the Skiplist.
//
func (l *Skiplist) Len() int {
	return l.cnt
}

// ElementN returns the Element at position pos in the skiplist, in O(log(index)) time.
// If no such entry exists, nil is returned.
//
func (l *Skiplist) ElementN(index int) *Element {
	if index >= l.cnt {
		return nil
	}
	prev := l.prevsN(index)
	return prev[0].link.to
}

// Function grow increments the list count and increment the number of
// levels on power-of-two counts.
//
func (l *Skiplist) grow() {
	l.cnt++
	if l.cnt&(l.cnt-1) == 0 {
		l.links = append(l.links, link{nil, l.cnt})
		l.prev = append(l.prev, prev{})
	}
}

type prev struct {
	link *link
	pos  int
}

// Return the previous links to modify, and the insertion position.
//
func (l *Skiplist) prevs(key interface{}, s float64) ([]prev, int) {
	levels := len(l.links)
	prev := l.prev
	links := &l.links
	pos := -1
	for level := levels - 1; level >= 0; level-- {
		// Find predecessor link at this level
		for (*links)[level].to != nil && ((*links)[level].to.score < s || (*links)[level].to.score == s && l.less((*links)[level].to.key, key)) {
			pos += (*links)[level].width
			links = &(*links)[level].to.links
		}
		prev[level].pos = pos
		prev[level].link = &(*links)[level]
	}
	pos++
	return prev, pos
}

// Return the previous links to modify, by index
//
func (l *Skiplist) prevsN(index int) []prev {
	levels := len(l.links)
	prev := l.prev
	links := &l.links
	pos := 0
	for level := levels - 1; level >= 0; level-- {
		// Find predecessor link at this level
		for (*links)[level].to != nil && (pos+(*links)[level].width <= index) {
			pos = pos + (*links)[level].width
			links = &(*links)[level].to.links
		}
		prev[level].pos = pos
		prev[level].link = &(*links)[level]
	}
	return prev
}

// Function randLevels returns a value from N from [0..limit-1] with probability
// 2^{-n-1}, except the last value is twice as likely.
//
func (l *Skiplist) randLevels(max int) int {
	levels := 1
	for r := l.rng.Int63(); 0 == r&1; r >>= 1 {
		levels++
	}
	if levels > max {
		return max
	}
	return levels
}

// Function shrink decrements the list count and decrement the number
// of levels on power-of-two counts.
//
func (l *Skiplist) shrink() {
	if l.cnt&(l.cnt-1) == 0 {
		l.links = l.links[:len(l.links)-1]
		l.prev = l.prev[:len(l.prev)-1]
	}
	l.cnt--
}

// Function String prints only the key/value pairs in the skip list.
//
func (l *Skiplist) String() string {
	s := append([]byte{}, "{"...)
	for n := l.links[0].to; n != nil; n = n.links[0].to {
		s = append(s, (n.String() + " ")...)
	}
	s[len(s)-1] = '}'
	return string(s)
}

// The SlowKey interface allows externally-defined types to be used 
// as keys.  An a.Less(b) call should return true iff a < b.
//
type SlowKey interface {
	Less(interface{}) bool
}

// The FastKey interface allows externally-defined types to be used
// as keys, efficiently.  An a.Less(b) call should return true iff a < b.
// key.Score() must increase monotonically as key increases.
//
type FastKey interface {
	Less(interface{}) bool
	Score() float64
}

// Function lessFn returns the comparison function corresponding to the key type.
//
func lessFn(key interface{}) func(a, b interface{}) bool {
	switch key.(type) {

	// Interface types come first to override builtin types when
	// the interface is present.

	case FastKey, SlowKey:
		return func(a, b interface{}) bool { return a.(SlowKey).Less(b) }

		// Support builtin types.

	case float32:
		return func(a, b interface{}) bool { return a.(float32) < b.(float32) }
	case float64:
		return func(a, b interface{}) bool { return a.(float64) < b.(float64) }
	case int:
		return func(a, b interface{}) bool { return a.(int) < b.(int) }
	case int16:
		return func(a, b interface{}) bool { return a.(int16) < b.(int16) }
	case int32:
		return func(a, b interface{}) bool { return a.(int32) < b.(int32) }
	case int64:
		return func(a, b interface{}) bool { return a.(int64) < b.(int64) }
	case int8:
		return func(a, b interface{}) bool { return a.(int8) < b.(int8) }
	case string:
		return func(a, b interface{}) bool { return a.(string) < b.(string) }
	case uint:
		return func(a, b interface{}) bool { return a.(uint) < b.(uint) }
	case uint16:
		return func(a, b interface{}) bool { return a.(uint16) < b.(uint16) }
	case uint32:
		return func(a, b interface{}) bool { return a.(uint32) < b.(uint32) }
	case uint64:
		return func(a, b interface{}) bool { return a.(uint64) < b.(uint64) }
	case uint8:
		return func(a, b interface{}) bool { return a.(uint8) < b.(uint8) }
	case uintptr:
		return func(a, b interface{}) bool { return a.(uintptr) < b.(uintptr) }

		// Support go-supplied type that are likely to be used as keys.

	case []byte:
		return func(a, b interface{}) bool { return bytes.Compare(a.([]byte), b.([]byte)) < 0 }
	}
	panic(fmt.Sprintf("skiplist: %T not supported.  Consider adding the SlowKey interface.", key))
}

// Function lessFn returns the comparison function corresponding to the key type.
//
func greaterFn(key interface{}, descending bool) func(a, b interface{}) bool {
	switch key.(type) {

	// Interface types come first to override builtin types when
	// the interface is present.

	case FastKey, SlowKey:
		return func(a, b interface{}) bool { return b.(SlowKey).Less(a) }

		// Support builtin types.

	case float32:
		return func(a, b interface{}) bool { return b.(float32) < a.(float32) }
	case float64:
		return func(a, b interface{}) bool { return b.(float64) < a.(float64) }
	case int:
		return func(a, b interface{}) bool { return b.(int) < a.(int) }
	case int16:
		return func(a, b interface{}) bool { return b.(int16) < a.(int16) }
	case int32:
		return func(a, b interface{}) bool { return b.(int32) < a.(int32) }
	case int64:
		return func(a, b interface{}) bool { return b.(int64) < a.(int64) }
	case int8:
		return func(a, b interface{}) bool { return b.(int8) < a.(int8) }
	case string:
		return func(a, b interface{}) bool { return b.(string) < a.(string) }
	case uint:
		return func(a, b interface{}) bool { return b.(uint) < a.(uint) }
	case uint16:
		return func(a, b interface{}) bool { return b.(uint16) < a.(uint16) }
	case uint32:
		return func(a, b interface{}) bool { return b.(uint32) < a.(uint32) }
	case uint64:
		return func(a, b interface{}) bool { return b.(uint64) < a.(uint64) }
	case uint8:
		return func(a, b interface{}) bool { return b.(uint8) < a.(uint8) }
	case uintptr:
		return func(a, b interface{}) bool { return b.(uintptr) < a.(uintptr) }

		// Support go-supplied type that are likely to be used as keys.

	case []byte:
		return func(a, b interface{}) bool { return bytes.Compare(b.([]byte), a.([]byte)) < 0 }
	}
	panic(fmt.Sprintf("skiplist: %T not supported.  Consider adding the SlowKey interface.", key))
}
