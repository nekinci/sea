package main

type Stack[T comparable] struct {
	values []T
	size   int
}

func (s *Stack[T]) Pop() T {
	if s.size > 0 {
		value := s.values[s.size-1]
		s.values = s.values[0 : s.size-1]
		s.size--
		return value
	}

	panic("stack is empty")
}

func (s *Stack[T]) Push(t T) {
	if s.values == nil {
		s.values = make([]T, 0)
	}
	s.size++
	s.values = append(s.values, t)
}

func (s *Stack[T]) Len() int {
	return s.size
}

func (s *Stack[T]) Empty() bool {
	return s.size == 0
}

func (s *Stack[T]) Reverse() *Stack[T] {
	s2 := &Stack[T]{}

	for !s.Empty() {
		s2.Push(s.Pop())
	}

	return s2
}
