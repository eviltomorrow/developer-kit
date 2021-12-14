package main

type (
	Stack struct {
		top    *node
		length int
	}
	node struct {
		value byte
		prev  *node
	}
)

var (
	dictDot = map[byte]struct{}{
		'"':  {},
		'\'': {},
	}
)

// Create a new stack
func NewStack() *Stack {
	return &Stack{nil, 0}
}

// Return the number of items in the stack
func (s *Stack) Len() int {
	return s.length
}

// View the top item on the stack
func (s *Stack) Peek() byte {
	if s.length == 0 {
		return 255
	}
	return s.top.value
}

// Pop the top item of the stack and return it
func (s *Stack) Pop() byte {
	if s.length == 0 {
		return 255
	}
	n := s.top
	s.top = n.prev
	s.length--
	return n.value
}

// Push a value onto the top of the stack
func (s *Stack) Push(value byte) {
	n := &node{value, s.top}
	s.top = n
	s.length++
}
