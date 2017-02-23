package vm

import (
	"context"
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
	return fmt.Sprintf("unknown operation '%s' on '%s'", e.Op, e.Class.FullName())
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

func (vm *VM) checkArity(m *value.Method, args []value.Value) error {
	if len(args) < m.Signature.Required {
		return &ErrArityMismatch{Got: len(args), Need: m.Signature.Required}
	}

	return nil
}

func (vm *VM) callN(ctx context.Context, recv value.Value, args []value.Value, op string) (value.Value, error) {
	fmt.Printf("op: %s\n", op)

	if t, ok := recv.Class(vm).LookupMethod(op); ok {
		if err := vm.checkArity(t, args); err != nil {
			return nil, err
		}

		return t.Func(ctx, vm, recv, args)
	}

	return nil, errors.WithStack(&ErrUnknownOp{Op: op, Class: recv.Class(vm)})
}

func (vm *VM) invoke(ctx context.Context, args []value.Value) (value.Value, error) {
	l := args[0].(*value.Lambda)

	if len(args)-1 < l.Args {
		return nil, &ErrArityMismatch{Got: len(args), Need: l.Args}
	}

	sub := value.ExecuteContext{
		Code: l.Code,
		Refs: l.Refs,
		Self: l.Self,
		Args: args[1:],
	}

	return vm.ExecuteContext(ctx, sub)
}

func (vm *VM) InvokeLambda(ctx context.Context, l *value.Lambda, args []value.Value) (value.Value, error) {
	if len(args) < l.Args {
		return nil, &ErrArityMismatch{Got: len(args), Need: l.Args}
	}

	sub := value.ExecuteContext{
		Code: l.Code,
		Refs: l.Refs,
		Self: l.Self,
		Args: args,
	}

	return vm.ExecuteContext(ctx, sub)
}
