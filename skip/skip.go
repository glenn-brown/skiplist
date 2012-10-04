// Package skip implements a skip list.  A skip.List stores
// key/value pairs sorted by key.  Values can be referenced by key or
// position.  Operations require O(log(N)) time and space.  The
// skiplist requires O(N) space.
package skip

import (
	"math/rand"
)

type Lesser interface {
	Less (Lesser) bool
}

type List struct {
	links []link
	cnt int
}
type link struct {
	to *node
	width int
}
type node struct {
	links []link
	key Lesser
	val interface{}
}

func (l *List) Insert(key Lesser, val interface{}) *List{
	l.grow()
	level := len(l.links)-1
	insert(&l.links[level].to, level, &node{[]link{}, key, val})
	return l
}
func insert(p **node, level int, nu *node) (width int, stop bool) {
	// Walk to the correct insertion location.
	for *p != nil && (*p).key.Less(nu.key) {
		width += (*p).links[level].width
		p = &(*p).links[level].to
	}
	// At the bottom level, simply link in the node
	if level == 0 {
		nu.links[0].to = *p
		nu.links[0].width = 1
		return 1, false
	}
	// Link in the new node at the lower levels.
	width, stop = insert(p, level-1, nu)
	// Don't link in the node if lower levels have stopped linking,
	// or if we randomly chose to stop linking.
	if stop || rand.Intn(2) == 0 {
		// Update the link width to reflect the insertion.
		(*p).links[level].width += 1
		return 0, true
	}
	// Link in the node on this level.
	fullWidth := (*p).links[level].width + 1
	(*p).links[level].width = width
	nu.links[level].width = fullWidth - width
	nu.links[level].to = nu
	return width+1, false
}

// Remove the youngest key/value pair associated with key from the skiplist, if a match exists.
func (l *List) Remove(key Lesser) (val interface{}, ok bool) {
	level := len(l.links)-1
	removed, ok := remove(&l.links[level].to, level-1, key)
	if !ok {
		return nil, false
	}
	return removed.val, true
}
func remove(p **node, level int, key Lesser) (removed *node, ok bool) {
	for *p != nil && (*p).key.Less(key) {
		p = &(*p).links[level].to
	}
	if level == 0 {
		if key.Less((*p).key) {
			return nil, false
		}
		removed = (*p).links[level].to
		(*p).links[0].to = removed.links[0].to
		return removed, true
	}
	removed, ok = remove(p, level-1, key)
	if ok {
		if (*p).links[level].to == removed {
			(*p).links[level].to = removed.links[level].to
			(*p).links[level].width += removed.links[level].width - 1
		}
		(*p).links[level].width -= 1;
	}
	return removed, ok
}

// RemoveIndex returns and removes the key,value pair stored at pos, in O(N) time.
func (l *List) RemoveN(pos int) (key Lesser, val interface{}, ok bool) {
	level := len(l.links)-1
	removed, ok := removeN(&l.links[level].to, level, pos)
	if !ok {
		return nil, nil, false
	}
	return removed.key, removed.val, true
}
func removeN (p **node, level int, pos int) (removed *node, ok bool) {
	for *p != nil && pos > (*p).links[level].width {
		pos -= (*p).links[level].width
		p = &(*p).links[level].to
	}
	if level == 0 {
		if pos != (*p).links[level].width {
			return nil, false
		}
		removed = (*p).links[level].to
		(*p).links[0].to = removed.links[0].to
		return removed, true
	}
	removed, ok = removeN(p, level-1, pos)
	if (*p).links[level].to == removed {
		(*p).links[level].to = removed.links[level].to
		(*p).links[level].width += removed.links[level].width - 1
	} else if ok {
		(*p).links[level].width -= 1;
	}
	return removed, ok
}

// Find returns the youngest value associated with key in the skiplist.
func (l *List) Peek (key Lesser) (val interface{}, ok bool) {
	level := len(l.links)-1
	n := l.links[level].to;
	for level >= 0 {
		for n!=nil && n.key.Less(key) {
			n = n.links[level].to
		}
		level--
	}
	if key.Less(n.key) {
		return nil, false
	}
	return n.val, true
}

// ValN returns the youngest key,value pair stored at pos.
func (l *List) PeekN(pos int) (key Lesser, val interface{}, ok bool) {
	pos++
	level := len(l.links)-1
	n := l.links[level].to;
	for level >= 0 {
		for n!=nil && pos >= n.links[level].width {
			pos -= n.links[level].width
			n = n.links[level].to
		}
		level--
	}
	if pos != 0 {
		return nil, nil, false
	}
	return n.key, n.val, true
}

// Len returns the number of elements in the List.
func (l *List) Len() int {
	return l.cnt
}

// Increment the list count and increment the number of levels on power-of-two counts.
func (l *List) grow() {
	l.cnt++
	if l.cnt & (l.cnt-1) == 0 {
		l.links = append(l.links, link{nil,l.cnt+1})
	}
}

// Decrement the list count and decrement the number of levels on power-of-two counts.
func (l *List) shrink() {
	if l.cnt & (l.cnt-1) == 0 {
		l.links = l.links[:len(l.links)-1]
	}
	l.cnt--
}
