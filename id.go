package goui

// ID is the identifier of a [Widget].
type ID interface {
	privateImplementsID() // unexported to prevent external implementations
}

// valueID is an ID implementation that uses a comparable value.
type valueID[T comparable] struct {
	value T
}

func (valueID[T]) privateImplementsID() {}

// ValueID creates an ID from a comparable value.
func ValueID[T comparable](value T) ID {
	return valueID[T]{value: value}
}
