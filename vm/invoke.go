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

func (vm *VM) callN(recv value.Value, args []value.Value, op string) (value.Value, error) {
	if t, ok := recv.Type(vm).Methods[op]; ok {
		return t.F(vm, recv, args)
	}

	return nil, errors.WithStack(&ErrUnknownOp{Op: op, Type: recv.Type(vm)})
}

func (vm *VM) invoke(ctx ExecuteContext, args []value.Value) (value.Value, error) {
	l := args[0].(*value.Lambda)

	sub := ExecuteContext{
		Sp:   ctx.Sp + ctx.Code.NumRegs,
		Code: l.Code,
		Refs: l.Refs,
	}

	// TODO use overlapping call args with locals in invoked lambda
	// rather than copy them

	for i, v := range args[1:] {
		vm.reg[sub.Sp+i] = v
	}

	return vm.ExecuteContext(sub)
}
