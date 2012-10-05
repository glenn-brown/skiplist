// Package skiplist implements a key- and position-addressible skip
// list with automatic level adjustment.  Insert, Remove, and Peek
// operations require O(log(N)) time and space.  The skiplist requires
// O(N) space.
//
// A skiplist is linked at multiple levels.  The bottom level (L0) is
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
//	
package skiplist

import (
	"fmt"
	"math/rand"
)

type Skiplist struct {
	cnt   int
	less  func(a, b interface{}) bool
	links []link
	rng   *rand.Rand
}
type link struct {
	to    *node
	width int
}
type node struct {
	links []link
	key   interface{}
	val   interface{}
}

// New returns a new skiplist.
// Function less must return true iff key a is less than key b.
// 
func New(less func(a, b interface{}) bool, r *rand.Rand) *Skiplist {
	if r == nil {
		r = rand.New(rand.NewSource(42))
	}
	return &Skiplist{0, less, []link{}, r}
}

// Insert inserts a {key,val} pair into the skip list.
func (l *Skiplist) Insert(key interface{}, val interface{}) *Skiplist {
	l.grow()
	level := len(l.links) - 1
	l.insert(&l.links, level, &node{[]link{}, key, val})
	return l
}

func (l *Skiplist) insert(links *[]link, level int, nu *node) (lwidth int, stopped bool) {
	// Walk to the correct insertion location
	for (*links)[level].to != nil && l.less((*links)[level].to.key, nu.key) {
		lwidth += (*links)[level].width
		links = &(*links)[level].to.links
	}
	// At the bottom level, simply link in the node
	if level == 0 {
		nu.links = append(nu.links, link{(*links)[0].to, 1})
		(*links)[0].to = nu
		return lwidth, false
	}
	// Link in the new node at the lower levels.
	rwidth, stopped := l.insert(links, level-1, nu)
	// Don't link in the node if lower levels have stopped linking,
	// or if we randomly chose to stop linking.
	if stopped || l.rng.Intn(2) < 1 {
		(*links)[level].width += 1
		return 0, true
	}
	// Link in the node on this level.
	nu.links = append(nu.links, link{(*links)[level].to, (*links)[level].width - rwidth})
	(*links)[level].to = nu
	(*links)[level].width = rwidth + 1
	return lwidth + rwidth, false
}

// Remove removes the youngest key/value pair associated with key from
// the skiplist, if a match exists.
func (l *Skiplist) Remove(key interface{}) (val interface{}, ok bool) {
	removed := l.remove(&l.links, len(l.links) - 1, key)
	if removed == nil {
		return nil, false
	}
	l.cnt--
	return removed.val, true
}
func (s *Skiplist) remove(links *[]link, level int, key interface{}) (removed *node) {
	// Walk to the correct deletion location
	for (*links)[level].to != nil && s.less((*links)[level].to.key, key) {
		links = &(*links)[level].to.links
	}
	// For level 0, simply remove the found node.
	if level == 0 {
		n := (*links)[0].to
		if n == nil || s.less(key, n.key) {
			return nil
		}
		(*links)[0].to = n.links[0].to
		// width is already 1 for level 0
		return n
	}
	// For other levels, recur.
	removed = s.remove(links, level-1, key)
	if (removed != nil) {
		if removed == (*links)[level].to {
			(*links)[level].to = removed.links[level].to
			(*links)[level].width += removed.links[level].width - 1
		} else {
			(*links)[level].width -= 1
		}
	}
	return removed
}

// RemoveIndex returns and removes the key,value pair stored at pos, in O(N) time.
func (l *Skiplist) RemoveN(pos int) (key interface{}, val interface{}, ok bool) {
	pos++
	if pos > l.cnt {
		return nil, nil, false
	}
	level := len(l.links) - 1
	removed := removeN(&l.links, level, pos)
	if removed == nil {
		return nil, nil, false
	}
	l.cnt--
	return removed.key, removed.val, true
}
func removeN(links *[]link, level int, pos int) (removed *node) {
	// Walk to the correct deletion location
	for (*links)[level].width < pos {
		pos -= (*links)[level].width
		links = &(*links)[level].to.links
	}
	// For level 0, simply remove the found node.
	if level == 0 {
		removed = (*links)[0].to
		(*links)[0].to = removed.links[0].to;
		return removed;
	}
	/* For higher levels, recur. */
	removed = removeN(links, level-1, pos)
	if removed == (*links)[level].to {
		// Unlink node at this level.
		(*links)[level].to = removed.links[level].to
		(*links)[level].width += removed.links[level].width - 1
	} else {
		// Account for node not linked at this level.
		(*links)[level].width--
	}
	return removed
}

// Peek returns the youngest value associated with key in the skiplist, without modifying the list.
func (l *Skiplist) Peek(key interface{}) (val interface{}, ok bool) {
	p := &l.links
	level := len(l.links) - 1
	for level >= 0 {
		for (*p)[level].to != nil && l.less((*p)[level].to.key, key) {
			p = &(*p)[level].to.links
		}
		level--
	}
	if (*p)[0].to == nil || l.less(key, (*p)[0].to.key) {
		return nil, false
	}
	return (*p)[0].to.val, true
}

// Len returns the number of elements in the Skiplist.
func (l *Skiplist) Len() int {
	return l.cnt
}

// Find returns the youngest value associated with key in the skiplist.
func (l *Skiplist) PeekN(pos int) (key interface{}, val interface{}, ok bool) {
	pos++
	p := &l.links
	level := len(l.links) - 1
	for level >= 0 {
		for (*p)[level].to != nil && (*p)[level].width < pos {
			pos -= (*p)[level].width
			p = &(*p)[level].to.links
		}
		level--
	}
	if (*p)[0].to == nil || pos < (*p)[0].width {
		return nil, nil, false
	}
	return (*p)[0].to.key, (*p)[0].to.val, true
}

// Function grow increments the list count and increment the number of
// levels on power-of-two counts.
func (l *Skiplist) grow() {
	l.cnt++
	if l.cnt&(l.cnt-1) == 0 {
		l.links = append(l.links, link{nil, l.cnt})
	}
}

// Function shrink decrements the list count and decrement the number
// of levels on power-of-two counts.
func (l *Skiplist) shrink() {
	if l.cnt&(l.cnt-1) == 0 {
		l.links = l.links[:len(l.links)-1]
	}
	l.cnt--
}

// Function String prints only the key/value pairs in the skip list.
func (l *Skiplist) String() string {
	s := append([]byte{}, "{"...)
	for n := l.links[0].to; n != nil; n = n.links[0].to {
		s = append(s, fmt.Sprint(n.key, ":", n.val, " ")...)
	}
	s[len(s)-1] = '}'
	return string(s)
}
