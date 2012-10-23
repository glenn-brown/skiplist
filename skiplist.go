// Copyright 2012 by the Skiplist Authors

// Package skiplist implements an *indexable* ordered map/multimap
//
// This skip list has special features:
// It supports position-index addressing.
// It can act as a map or as a multimap.
// It automatically adjusts its depth.
// It mimics Go's container/list interface where possible.
// It automatically supports integer, float, and []byte keys.
// It supports external key types via the FastKey and SlowKey interfaces.
//
// Set, Get, Insert, and Remove, operations all require O(log(N)) time or less.
// The skiplist requires O(N) space.
//
// To efficiently iterate over the list (where s is a *Skiplist):
//   for e := l.Front(); e != nil; e = e.Next() {
//  	// do something with e.Value and/or e.Key()
//   }
// Pop the first element in the list with l.RemoveN(0).
// Pop the last with l.RemoveN(l.Len()-1).
//
// To use the skiplist as a Map, mapping each key to a single value, simply avoid the Insert() method.
//	
// To use the skiplist as a Multimap, use Insert() instead of Set().
// To efficiently iterate over all values for a single key:
//   for e := l.Get(key); e != nil && e.Key() == key; e = e.Next() {
//     ;
//   }
//
package skiplist

import (
	"bytes"
	"fmt"
	"math/rand"
)

// A Skiplist is linked at multiple levels.  The bottom level (L0) is
// a sorted linked list of entries, and each link has a link at the
// next higher level added with probability P at insertion.  Since
// this is a position-addressable skip-list, each link has an
// associated 'width' specifying the number of nodes it skips, so
// nodes can also be referenced by position.
//
// For example, here is a 24-entry skip list:
//   L4 |---------------------------------------------------------------------->/
//   L3 |------------------------------------------->|------------------------->/
//   L2 |---------->|---------->|---------->|------->|---------------->|---->|->/
//   L1 |---------->|---------->|---------->|->|---->|->|->|->|------->|->|->|->/
//   L0 |->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->|->/
//         0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  1  1  1  1  1  1  1  
//         0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f  0  1  2  3  4  5  6
type Skiplist struct {
	cnt   int
	less  func(a, b interface{}) bool
	links []link
	prev  []prev
	rng   *rand.Rand
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

// Key returns the key used to insert the value in the list element.
//
func (e *Element) Key() interface{} { return e.key }

// Next returns the next-greater list element or nil.
//
func (e *Element) Next() *Element { return e.links[0].to }

// String returns a Key:Value string representation.
//
func (e *Element) String() string { return fmt.Sprintf("%v:%v", e.Key(), e.Value) }

// New returns a new skiplist in O(1) time.
// Function less must return true iff key a is less than key b.
// The list will be sorted from least to greatest.
// R is the random number generator to use or nil.
//
func New(r *rand.Rand) *Skiplist {
	if r == nil {
		r = rand.New(rand.NewSource(42))
	}
	nu := &Skiplist{0, nil, []link{}, []prev{}, r}
	// Arrange to set the less function the first time it it called.
	// We can't do it here because we do not yet know the key type.
	nu.less = func(a, b interface{}) bool {
		nu.less = lessFn(a)
		return nu.less(a, b)
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

// Insert a {key,value} pair in the skiplist, optionally replacing the yougest previous entry.
//
func (l *Skiplist) insert(key interface{}, value interface{}, replace bool) *Skiplist {
	l.grow()
	s := score(key)
	prev, pos := l.prevs(key, s)
	next := prev[0].link.to
	if replace && (s != next.score || s == next.score && (l.less(key, next.key) || l.less(next.key, key))) {
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

// Insert a {key,value} pair into the skip list in O(log(N)) time, replacing the youngest entry
// for key, if any.
//
func (l *Skiplist) Set(key interface{}, value interface{}) *Skiplist {
	return l.insert(key, value, true)
}

func (l *Skiplist) remove(prev []prev, elem *Element) *Element {
	// At the bottom level, simply unlink the element.
	prev[0].link.to = elem.links[0].to
	// Unlink any higher linked levels.
	level := 1
	levels := len(l.links)
	for ; level < levels && prev[level].link.to == elem; level++ {
		prev[level].link.to = elem.links[level].to
		prev[level].link.width += elem.links[level].width
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
	s := score(key)
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
	panic ("TODO")
}

// RemoveN removes any element at position pos in O(log(N)) time,
// returning it or nil.
//
func (l *Skiplist) RemoveN(index int) *Element {
	if index >= l.cnt {
		return nil
	}
	prev := l.prevsN(index)
	elem := prev[0].link.to
	for level := range l.links {
		if level < len(elem.links) {
			if level == 0 {
				// At the bottom level, simply unlink the element.
				prev[level].link.to = elem.links[level].to
				continue
			}
			prev[level].link.to = elem.links[level].to
			prev[level].link.width += elem.links[level].width - 1
			continue
		}
		prev[level].link.width -= 1
	}
	l.shrink()
	return elem
}

// Find returns the youngest element inserted with key in the
// skiplist, without modifying the list, in O(log(N)) time.
// If there is no match, nil is returned.
// It also returns the current position of the found element, or -1.
//
func (l *Skiplist) Find(key interface{}) (e *Element, pos int) {
	s := score(key)
	prev, pos := l.prevs(key, s)
	elem := prev[0].link.to
	if elem == nil || s < elem.score || s == elem.score && l.less(key, elem.key) {
		return nil, -1
	}
	return elem, pos
}

// Len returns the number of elements in the Skiplist.
//
func (l *Skiplist) Len() int {
	return l.cnt
}

// FindN returns the Element at position pos in the skiplist, in O(log(index)) time.
// If no such entry exists, nil is returned.
//
func (l *Skiplist) FindN(index int) *Element {
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
	for l.rng.Int63()&0x8000 != 0 {
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

// Any type implementing the SlowKey interface may be used as a key,
// but the FastKey interface is faster.
//
type SlowKey interface {
	Less(interface{}) bool
}

// Any type implementing the FastKey interface may be used as a key.
// a<b => Score(a)<=Score(b)
//
type FastKey interface {
	Less(interface{}) bool
	Score() float64
}

// Function lessFn returns the comparison function corresponding to the key type.
//
func lessFn(key interface{}) func(a,b interface{})bool {
	switch key.(type) {

		// Support builtin types.
		
	case float32 :return func(a,b interface{})bool{return a.(float32)<b.(float32)}
	case float64 :return func(a,b interface{})bool{return a.(float64)<b.(float64)}
	case int     :return func(a,b interface{})bool{return a.(int    )<b.(int    )}
	case int16   :return func(a,b interface{})bool{return a.(int16  )<b.(int16  )}
	case int32   :return func(a,b interface{})bool{return a.(int32  )<b.(int32  )}
	case int64   :return func(a,b interface{})bool{return a.(int64  )<b.(int64  )}
	case int8    :return func(a,b interface{})bool{return a.(int8   )<b.(int8   )}
	case string  :return func(a,b interface{})bool{return a.(string )<b.(string )}
	case uint    :return func(a,b interface{})bool{return a.(uint   )<b.(uint   )}
	case uint16  :return func(a,b interface{})bool{return a.(uint16 )<b.(uint16 )}
	case uint32  :return func(a,b interface{})bool{return a.(uint32 )<b.(uint32 )}
	case uint64  :return func(a,b interface{})bool{return a.(uint64 )<b.(uint64 )}
	case uint8   :return func(a,b interface{})bool{return a.(uint8  )<b.(uint8  )}
	case uintptr :return func(a,b interface{})bool{return a.(uintptr)<b.(uintptr)}

		// Support go-supplied type that are likely to be used as keys.
		
	case []byte  :return func(a,b interface{})bool{return bytes.Compare(a.([]byte),b.([]byte)) < 0}

		// Support types that implement the SlowKey and FastKey interfaces.
		
	case SlowKey, FastKey: return func(a,b interface{})bool{return a.(SlowKey).Less(b)}
	}
	panic ("skiplist: type T not supported.  Consider adding a Less() method.")
}
