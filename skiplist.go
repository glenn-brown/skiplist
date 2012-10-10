// Package skiplist implements an indexable ordered multimap.
//
// This skip list implementation is distinguished by supporting index
// addressing, allowing multiple values per key, automatically
// adjusting its depth, and mimicing Go's container/list interface
// where possible.
//
// Insert, Remove, and Find operations all require O(log(N)) time or less.
// The skiplist requires O(N) space.
//
// To efficiently iterate over the list (where s is a *Skiplist):
//   for e := s.Front(); e != nil; e = e.Next() {
//  	// do something with e.Value and/or e.Key()
//   }
// Pop the first element in the list with s.RemoveN(0).  Pop the last
// with s.RemoveN(s.Len()-1).
//
package skiplist

import (
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
	prev  []struct{		// Scratch structure
		link *link
		pos int
	}
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
	links []link
	key   interface{} // private to protect order
	Value interface{}
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
func New(less func(key1, key2 interface{}) bool, r *rand.Rand) *Skiplist {
	if r == nil {
		r = rand.New(rand.NewSource(42))
	}
	return &Skiplist{0, less, []link{}, []struct{link *link;pos int}{}, r}
}

// Return the first list element in O(1) time.
//
func (s *Skiplist) Front() *Element {
	if len(s.links) == 0 {
		return nil
	}
	return s.links[0].to
}

// Insert a {key,value} pair into the skip list in O(log(N)) time.
//
func (l *Skiplist) Insert(key interface{}, value interface{}) *Skiplist {
	l.grow()
	levels := len(l.links)
	// Create scratch space to store predecessor information.
	prev := l.prev
	// Compute elements preceding the insertion location at each level.
	pos := 0
	links := l.links
	for level := levels-1; level >= 0; level-- {
		ll := &links[level]
		// Find predecessor link at this level.
		for ll.to != nil && l.less(ll.to.key, key) {
			pos += ll.width
			links = ll.to.links
			ll = &links[level]
		}
		// Increment the width of the 
		ll.width += 1
		// Record the predecessor at this level and its position.
		prev[level].pos = pos
		prev[level].link = ll
	}
	// Set pos to the position of the new element.
	pos++
	// At the bottom level, simply link in the element
	nu := &Element{make([]link,1,2), key, value}
	nu.links[0] = link{prev[0].link.to, 1}
	prev[0].link.to = nu
	prev[0].link.width = 1
	// Link in the element at a random number of higher levels.
	for level:=1; level<levels && l.rng.Intn(2) < 1; level++ {
		end := prev[level].pos + prev[level].link.width
		nu.links = append(nu.links, link{prev[level].link.to, end - pos})
		prev[level].link.to = nu
		prev[level].link.width = pos - prev[level].pos
	}
	return l
}

// Remove the youngest Element associate with Key, if any, in O(log(N)) time.
// Return the removed element or nil.
//
func (l *Skiplist) Remove(key interface{}) *Element {
	levels := len(l.links)
	// Create scratch space to store predecessor information.
	prev := l.prev
	// Compute elements preceding the insertion location at each level.
	links := l.links
	for level := levels-1; level >= 0; level-- {
		ll := &links[level]
		// Find predecessor link at this level.
		for ll.to != nil && l.less(ll.to.key, key) {
			links = ll.to.links
			ll = &links[level]
		}
		// Record the predecessor at this level and its position.
		prev[level].link = ll
	}
	// Verify there is a matching entry to remove.
	elem := prev[0].link.to
	if elem == nil || l.less(key, elem.key) {
		return nil
	}
	// At the bottom level, simply unlink the element.
	prev[0].link.to = elem.links[0].to
	// Unlink any higher linked levels.
	level := 1
	for ; level<levels && prev[level].link.to == elem; level++ {
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

// RemoveN removes any element at position pos in O(log(N)) time,
// returning it or nil.
//
func (l *Skiplist) RemoveN(index int) *Element {
	if index >= l.cnt {
		return nil
	}
	levels := len(l.links)
	// Create scratch space to store predecessor information.
	prev := l.prev
	// Compute elements preceding the insertion location at each level.
	pos := -1
	links := l.links
	for level := levels-1; level >= 0; level-- {
		ll := &links[level]
		// Find predecessor link at this level.
		for ll.to != nil && pos + ll.width < index {
			pos += ll.width
			links = ll.to.links
			ll = &links[level]
		}
		// Record the predecessor at this level and its position.
		prev[level].pos = pos
		prev[level].link = ll
	}
	// Set pos to the position of the new element.
	pos++
	// Verify there is a matching entry to remove.
	elem := prev[0].link.to
	// At the bottom level, simply unlink the element.
	prev[0].link.to = elem.links[0].to
	// Unlink any higher linked levels.
	level := 1
	for ; level<levels && prev[level].link.to == elem; level++ {
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

// Find returns the youngest element inserted with key in the
// skiplist, without modifying the list, in O(log(N)) time.
// If there is no match, nil is returned.
// It also returns the current position of the found element, or -1.
//
func (l *Skiplist) Find(key interface{}) (e *Element, pos int) {
	levels := len(l.links)
	// Create scratch space to store predecessor information.
	prev := l.prev
	// Compute elements preceding the insertion location at each level.
	links := l.links
	for level := levels-1; level >= 0; level-- {
		ll := &links[level]
		// Find predecessor link at this level.
		for ll.to != nil && l.less(ll.to.key, key) {
			pos += ll.width
			links = ll.to.links
			ll = &links[level]
		}
		// Record the predecessor at this level and its position.
		prev[level].pos = pos
		prev[level].link = ll
	}
	// Set pos to the position of the new element.
	elem := prev[0].link.to
	if elem == nil || l.less(key, elem.key) {
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
	levels := len(l.links)
	// Compute elements preceding the insertion location at each level.
	pos := -1
	links := l.links
	var ll *link
	for level := levels-1; level >= 0; level-- {
		ll = &links[level]
		// Find predecessor link at this level.
		for ll.to != nil && pos + ll.width < index {
			pos += ll.width
			links = ll.to.links
			ll = &links[level]
		}
	}
	return ll.to
}

// Function grow increments the list count and increment the number of
// levels on power-of-two counts.
//
func (l *Skiplist) grow() {
	l.cnt++
	if l.cnt&(l.cnt-1) == 0 {
		l.links = append(l.links, link{nil, l.cnt})
		l.prev = append(l.prev, struct{link *link;pos int}{})
	}
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
