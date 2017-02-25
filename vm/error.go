package vm

import (
	"fmt"

	"github.com/evanphx/m13/value"
)

type ErrTypeError struct {
	got, expected string
}

func (e *ErrTypeError) Error() string {
	return fmt.Sprintf("TypeError: Needed a '%s', got a '%s'", e.expected, e.got)
}

func (vm *VM) TypeError(val value.Value, expected string) (value.Value, error) {
	got := val.Class(vm).GlobalName

	return nil, &ErrTypeError{got, expected}
}
