package vm

import (
	"fmt"

	"github.com/evanphx/m13/value"
	"github.com/pkg/errors"
)

var ErrUnknownOp = errors.New("unknown op")

type ErrArityMismatch struct {
	Got  int
	Need int
}

func (e *ErrArityMismatch) Error() string {
	return fmt.Sprintf("arity mismatch: expected %d, got %d", e.Need, e.Got)
}

func (vm *VM) ArgumentError(got, need int) (value.Value, error) {
	return nil, &ErrArityMismatch{Got: got, Need: need}
}

func (vm *VM) invokeOp(l, r value.Value, op string) (value.Value, error) {
	if t, ok := l.Type().Methods[op]; ok {
		return t.F(vm, l, []value.Value{r})
	}

	return nil, errors.Wrapf(ErrUnknownOp, "types (%T, %T)", l, r)
}

func (vm *VM) invokeN(recv value.Value, args []value.Value, op string) (value.Value, error) {
	if t, ok := recv.Type().Methods[op]; ok {
		return t.F(vm, recv, args)
	}

	return nil, errors.Wrapf(ErrUnknownOp, "types (%T)", recv)
}
