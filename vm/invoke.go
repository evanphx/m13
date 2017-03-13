package vm

import (
	"context"
	"fmt"

	"github.com/evanphx/m13/value"
	"github.com/pkg/errors"
)

type ErrArityMismatch struct {
	Name string
	Got  int
	Need int
}

func (e *ErrArityMismatch) Error() string {
	return fmt.Sprintf("arity mismatch on method '%s': expected %d, got %d", e.Name, e.Need, e.Got)
}

type ErrUnknownOp struct {
	Op    string
	Class *value.Class
}

func (e *ErrUnknownOp) Error() string {
	return fmt.Sprintf("unknown operation '%s' on '%s'", e.Op, e.Class.FullName())
}

func (vm *VM) ArgumentError(got, need int) (value.Value, error) {
	return nil, &ErrArityMismatch{Name: "unknown", Got: got, Need: need}
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
		return &ErrArityMismatch{Name: m.Name, Got: len(args), Need: m.Signature.Required}
	}

	return nil
}

func (vm *VM) callN(ctx context.Context, recv value.Value, args []value.Value, call *value.CallSite) (value.Value, error) {
	if t, ok := recv.Class(vm).LookupMethod(call.Name); ok {
		if err := vm.checkArity(t, args); err != nil {
			return nil, err
		}

		return t.Func(ctx, vm, recv, args)
	}

	return nil, errors.WithStack(&ErrUnknownOp{Op: call.Name, Class: recv.Class(vm)})
}

func (vm *VM) callKW(
	ctx context.Context,
	recv value.Value,
	pos []value.Value,
	kw []value.Value,
	call *value.CallSite,
) (value.Value, error) {
	if t, ok := recv.Class(vm).LookupMethod(call.Name); ok {
		got := len(pos) + len(kw)
		if got < t.Signature.Required {
			return nil, &ErrArityMismatch{Name: call.Name, Got: got, Need: t.Signature.Required}
		}

		args := make([]value.Value, t.Signature.Required)

		copy(args, pos)

		for i, name := range call.KWTable {
			for j, sym := range t.Signature.Args {
				if name == sym {
					args[j] = kw[i]
				}
			}
		}

		return t.Func(ctx, vm, recv, args)
	}

	return nil, errors.WithStack(&ErrUnknownOp{Op: call.Name, Class: recv.Class(vm)})
}

func (vm *VM) invoke(ctx context.Context, args []value.Value) (value.Value, error) {
	l := args[0].(*value.Lambda)

	if len(args)-1 < l.Args {
		return nil, &ErrArityMismatch{Name: "invoke", Got: len(args), Need: l.Args}
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
		return nil, &ErrArityMismatch{Name: "invoke", Got: len(args), Need: l.Args}
	}

	sub := value.ExecuteContext{
		Code: l.Code,
		Refs: l.Refs,
		Self: l.Self,
		Args: args,
	}

	return vm.ExecuteContext(ctx, sub)
}
