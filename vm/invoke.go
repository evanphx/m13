package vm

import (
	"fmt"

	"github.com/evanphx/m13/value"
	"github.com/pkg/errors"
)

type ErrArityMismatch struct {
	Got  int
	Need int
}

func (e *ErrArityMismatch) Error() string {
	return fmt.Sprintf("arity mismatch: expected %d, got %d", e.Need, e.Got)
}

type ErrUnknownOp struct {
	Op   string
	Type *value.Type
}

func (e *ErrUnknownOp) Error() string {
	return fmt.Sprintf("unknown operation '%s' on '%s", e.Op, e.Type.FullName())
}

func (vm *VM) ArgumentError(got, need int) (value.Value, error) {
	return nil, &ErrArityMismatch{Got: got, Need: need}
}

func (vm *VM) MustFindType(globalName string) *value.Type {
	t, ok := vm.registry.FindType(globalName)
	if !ok {
		panic(fmt.Sprintf("unknown type: %s", globalName))
	}

	return t
}

func (vm *VM) invokeOp(l, r value.Value, op string) (value.Value, error) {
	if t, ok := l.Type(vm).Methods[op]; ok {
		return t.F(vm, l, []value.Value{r})
	}

	return nil, errors.WithStack(&ErrUnknownOp{Op: op, Type: l.Type(vm)})
}

func (vm *VM) invokeN(recv value.Value, args []value.Value, op string) (value.Value, error) {
	if t, ok := recv.Type(vm).Methods[op]; ok {
		return t.F(vm, recv, args)
	}

	return nil, errors.WithStack(&ErrUnknownOp{Op: op, Type: recv.Type(vm)})
}
