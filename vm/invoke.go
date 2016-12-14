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
	Op    string
	Class *value.Class
}

func (e *ErrUnknownOp) Error() string {
	return fmt.Sprintf("unknown operation '%s' on '%s", e.Op, e.Class.FullName())
}

func (vm *VM) ArgumentError(got, need int) (value.Value, error) {
	return nil, &ErrArityMismatch{Got: got, Need: need}
}

func (vm *VM) MustFindClass(globalName string) *value.Class {
	t, ok := vm.registry.FindClass(globalName)
	if !ok {
		panic(fmt.Sprintf("unknown type: %s", globalName))
	}

	return t
}

func (vm *VM) callN(recv value.Value, args []value.Value, op string) (value.Value, error) {
	if t, ok := recv.Class(vm).Methods[op]; ok {
		return t.F(vm, recv, args)
	}

	return nil, errors.WithStack(&ErrUnknownOp{Op: op, Class: recv.Class(vm)})
}

func (vm *VM) invoke(ctx value.ExecuteContext, args []value.Value) (value.Value, error) {
	l := args[0].(*value.Lambda)

	sub := value.ExecuteContext{
		Code: l.Code,
		Refs: l.Refs,
		Args: args[1:],
	}

	return vm.ExecuteContext(sub)
}
