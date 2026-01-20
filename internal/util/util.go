package util

import (
	"github.com/mkch/gg"
)

// LookupParent searches the element and its ancestors for an element
// that satisfies the given predicate.
// The found result of predicate indicates whether it is satisfied.
// The first value returned by predicate is returned when found.
// If no such element is found, the zero value of T and false are returned.
func LookupParent[
	Element interface {
		comparable
		Parent() Element
	},
	Ret any](element Element, predicate func(Element) (v Ret, found bool)) (ret Ret, found bool) {
	for elem := element; elem != gg.Zero[Element](); elem = elem.Parent() {
		if t, ok := predicate(elem); ok {
			return t, ok
		}
	}
	return
}

// LookupChild searches the element and its descendants for an element
// that satisfies the given predicate.
// The predicate function should return:
//   - satisfied=true: the element satisfies the condition, value is returned.
//   - satisfied=false: the element does not satisfy, continue searching all descendants.
//
// If an element satisfies the predicate, its value and true are returned immediately.
// If no such element is found, the zero value of T and false are returned.
func LookupChild[
	Element interface {
		comparable
		NumChildren() int
		Child(i int) Element
	},
	Ret any](element Element, predicate func(Element) (value Ret, satisfied bool)) (v Ret, found bool) {
	if t, ok := predicate(element); ok {
		return t, true
	}
	for i := 0; i < element.NumChildren(); i++ {
		if t, ok := LookupChild(element.Child(i), predicate); ok {
			return t, ok
		}
	}
	return
}
