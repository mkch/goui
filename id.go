package goui

// ID is an opaque identifier only comparable for equality.
// IDs are process-local and must not be serialized or passed across process boundaries.
type ID interface {
	privateImplementsID() // unexported to prevent external implementations
}

// valueID is an ID implementation that uses a comparable value.
type valueID[T comparable] struct {
	value T
}

func (valueID[T]) privateImplementsID() {}

// ValueID returns an ID backed by a comparable value.
// Two IDs created with ValueID are equal when their underlying values are equal; otherwise they are not.
// Do not use pointers to zero-sized types for T, as their equality is undefined according to Go spec.
func ValueID[T comparable](value T) ID {
	return valueID[T]{value: value}
}

// UniqueID creates and returns a process-wide unique ID.
func UniqueID() ID {
	// Uniqueness guarantees:
	// #1. Type unique is local and unexported, only this function can create a value of this type.
	// #2. As long as a unique value's address is still in use, new(unique) will not return the same address.
	type unique int
	return ValueID(new(unique))
}
