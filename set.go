package imdb

import (
	"fmt"
	"strings"
)

type void struct{}

type set struct {
	elements map[interface{}]void
}

// NewSet returns an empty set.
func NewSet() *set {
	return &set{make(map[interface{}]void)}
}

// Len returns number of elements in the set.
func (s *set) Len() int {
	return len(s.elements)
}

// Has checks if set has the element.
func (s *set) Has(e interface{}) bool {
	_, ok := s.elements[e]
	return ok
}

// Add adds an element to the set.
func (s *set) Add(e interface{}) {
	s.elements[e] = void{}
}

// Remove removes an element from the set.
func (s *set) Remove(e interface{}) {
	delete(s.elements, e)
}

// All returns all elements of the set.
func (s *set) All() []interface{} {
	elements := make([]interface{}, 0, len(s.elements))
	for element := range s.elements {
		elements = append(elements, element)
	}
	return elements
}

// String returns string representation of the set.
func (s *set) String() string {
	arr := make([]string, 0, len(s.elements))
	for element := range s.elements {
		arr = append(arr, fmt.Sprint(element))
	}
	return fmt.Sprintf("{%s}", strings.Join(arr, ", "))
}
