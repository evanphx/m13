package vm

import (
	"github.com/evanphx/m13/value"
	"github.com/pkg/errors"
)

var ErrUnknownOp = errors.New("unknown op")

func (vm *VM) ArgumentError(got, need int) (value.Value, error) {
	return nil, nil
}

func (vm *VM) invokeOp(l, r value.Value, op string) (value.Value, error) {
	if t, ok := l.Type().Methods[op]; ok {
		return t.F(vm, l, []value.Value{r})
	}

	return nil, errors.Wrapf(ErrUnknownOp, "types (%T, %T)", l, r)
}
