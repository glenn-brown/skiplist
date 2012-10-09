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
// For example, a 64-entry skip list with values from 0x00 to
// 0x3f might look like this:
//   L5 --------------------------------------------------------------->/
//   L4 ---------------------------------->->-------------------------->/
//   L3 ------------------->>-->-------->->->-------------------------->/
//   L2 -->------------>->->>-->-------->->->-->--->->---------->------>/
//   L1 >->->-->-->-->>>->->>>->-------->->->-->->->->-->--->--->------>/
//   L0 >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>/
//      0000000000000000111111111111111122222222222222223333333333333333
//      0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef

type Skiplist struct {
	cnt   int
	less  func(a, b interface{}) bool
	links []link
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
	return &Skiplist{0, less, []link{}, r}
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
	level := len(l.links) - 1
	l.insert(&l.links, level, &Element{[]link{}, key, value})
	return l
}

func (l *Skiplist) insert(links *[]link, level int, nu *Element) (lwidth int, stopped bool) {
	// Walk to the correct insertion location.
	for (*links)[level].to != nil && l.less((*links)[level].to.key, nu.key) {
		lwidth += (*links)[level].width
		links = &(*links)[level].to.links
	}
	// At the bottom level, simply link in the Element.
	if level == 0 {
		nu.links = append(nu.links, link{(*links)[0].to, 1})
		(*links)[0].to = nu
		return lwidth, false
	}
	// Link in the new Element at the lower levels.
	rwidth, stopped := l.insert(links, level-1, nu)
	// Don't link in the Element if lower levels have stopped linking,
	// or if we randomly chose to stop linking.
	if stopped || l.rng.Intn(2) < 1 {
		(*links)[level].width += 1
		return 0, true
	}
	// Link in the Element on this level.
	nu.links = append(nu.links, link{(*links)[level].to, (*links)[level].width - rwidth})
	(*links)[level].to = nu
	(*links)[level].width = rwidth + 1
	return lwidth + rwidth, false
}

// Remove the youngest Element associate with Key, if any, in O(1) time.
// Return the removed element or nil.
//
func (l *Skiplist) Remove(key interface{}) *Element {
	removed := l.remove(&l.links, len(l.links)-1, key)
	if removed != nil {
		l.shrink()
	}
	return removed
}
func (s *Skiplist) remove(links *[]link, level int, key interface{}) *Element {
	// Walk to the correct deletion location.
	for (*links)[level].to != nil && s.less((*links)[level].to.key, key) {
		links = &(*links)[level].to.links
	}
	// For level 0, simply remove the found Element.
	if level == 0 {
		n := (*links)[0].to
		if n == nil || s.less(key, n.key) {
			return nil
		}
		(*links)[0].to = n.links[0].to
		// width is already 1 for level 0.
		return n
	}
	// For other levels, recur.
	removed := s.remove(links, level-1, key)
	if removed != nil {
		if removed == (*links)[level].to {
			(*links)[level].to = removed.links[level].to
			(*links)[level].width += removed.links[level].width - 1
		} else {
			(*links)[level].width -= 1
		}
	}
	return removed
}

// RemoveN removes any element at position pos in O(log(pos)) time,
// returning it or nil.
//
func (l *Skiplist) RemoveN(pos int) *Element {
	pos++
	if pos > l.cnt {
		return nil
	}
	// Carefully the start-search level.  If we don't, the search
	// is technically O(log(l.cnt)) instead of O(log(pos)).
	levels := len(l.links)
	level := 0
	for ; level < levels-1; level++ {
		if pos < 1<<uint(level) {
			break
		}
	}
	removed := removeN(&l.links, level, pos)
	l.shrink()
	return removed
}

func removeN(links *[]link, level int, pos int) (removed *Element) {
	// Walk to the correct deletion location.
	for (*links)[level].width < pos {
		pos -= (*links)[level].width
		links = &(*links)[level].to.links
	}
	// For level 0, simply remove the found Element.
	if level == 0 {
		removed = (*links)[0].to
		(*links)[0].to = removed.links[0].to
		return removed
	}
	// For higher levels, recur.
	removed = removeN(links, level-1, pos)
	if removed == (*links)[level].to {
		// Unlink Element at this level.
		(*links)[level].to = removed.links[level].to
		(*links)[level].width += removed.links[level].width - 1
	} else {
		// Account for Element not linked at this level.
		(*links)[level].width--
	}
	return removed
}

// Find returns the youngest element inserted with key in the
// skiplist, without modifying the list, in O(log(N)) time.
// If there is no match, nil is returned.
// It also returns the current position of the found element, or -1.
//
func (l *Skiplist) Find(key interface{}) (e *Element, pos int) {
	p := &l.links
	level := len(l.links) - 1
	for level >= 0 {
		for (*p)[level].to != nil && l.less((*p)[level].to.key, key) {
			pos += (*p)[level].width
			p = &(*p)[level].to.links
		}
		level--
	}
	if (*p)[0].to == nil || l.less(key, (*p)[0].to.key) {
		return nil, -1
	}
	return (*p)[0].to, pos
}

// Len returns the number of elements in the Skiplist.
//
func (l *Skiplist) Len() int {
	return l.cnt
}

// FindN returns the Element at position pos in the skiplist, in O(log(pos)) time.
// If no such entry exists, nil is returned.
//
func (l *Skiplist) FindN(pos int) *Element {
	if pos >= l.cnt {
		return nil
	}
	pos++
	p := &l.links
	levels := len(l.links)
	// Carefully the start-search level.  If we don't, the search
	// is technically O(log(l.cnt)) instead of O(log(pos)).
	level := 0
	for ; level < levels-1; level++ {
		if pos < 1<<uint(level) {
			break
		}
	}
	for level >= 0 {
		for (*p)[level].to != nil && (*p)[level].width < pos {
			pos -= (*p)[level].width
			p = &(*p)[level].to.links
		}
		level--
	}
	if (*p)[0].to == nil || pos < (*p)[0].width {
		return nil
	}
	return (*p)[0].to
}

// Function grow increments the list count and increment the number of
// levels on power-of-two counts.
//
func (l *Skiplist) grow() {
	l.cnt++
	if l.cnt&(l.cnt-1) == 0 {
		l.links = append(l.links, link{nil, l.cnt})
	}
}

// Function shrink decrements the list count and decrement the number
// of levels on power-of-two counts.
//
func (l *Skiplist) shrink() {
	if l.cnt&(l.cnt-1) == 0 {
		l.links = l.links[:len(l.links)-1]
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
