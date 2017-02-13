package vm

import (
	"context"
	"fmt"

	"github.com/evanphx/m13/insn"
	"github.com/evanphx/m13/value"
)

type VM struct {
	registry *value.Registry
	reg      []value.Value
	top      int

	nil_   value.Value
	true_  value.Value
	false_ value.Value
}

func NewVM() (*VM, error) {
	reg := value.NewRegistry()
	nil_ := &value.Object{}
	nil_.SetClass(reg.NilClass)

	true_ := &value.Object{}
	true_.SetClass(reg.BoolClass)

	false_ := &value.Object{}
	false_.SetClass(reg.BoolClass)

	return &VM{
		registry: reg,
		reg:      make([]value.Value, 128),
		nil_:     nil_,
		true_:    true_,
		false_:   false_,
	}, nil
}

func (vm *VM) Registry() *value.Registry {
	return vm.registry
}

func (vm *VM) Reg(i int) value.Value {
	return vm.reg[i]
}

func (vm *VM) Class() *value.Class {
	return vm.registry.Class
}

func (vm *VM) ObjectClass() *value.Class {
	return vm.registry.Object
}

func (vm *VM) BoolClass() *value.Class {
	return vm.registry.BoolClass
}

func (vm *VM) I64Class() *value.Class {
	return vm.registry.I64Class
}

func (vm *VM) LambdaClass() *value.Class {
	return vm.registry.Lambda
}

func (vm *VM) StringClass() *value.Class {
	return vm.registry.Class
}

func (vm *VM) NewString(s string) *value.String {
	str := &value.String{
		String: s,
	}

	str.SetClass(vm.registry.String)

	return str
}

func (vm *VM) ListClass() *value.Class {
	return vm.registry.List
}

func (vm *VM) IOClass() *value.Class {
	return vm.registry.IO
}

func (vm *VM) Nil() value.Value {
	return vm.nil_
}

func (vm *VM) True() value.Value {
	return vm.true_
}

func (vm *VM) False() value.Value {
	return vm.false_
}

func (vm *VM) isTrue(v value.Value) bool {
	if v == nil || v == vm.nil_ || v == vm.false_ {
		return false
	}

	if b, ok := v.(value.Bool); ok {
		return bool(b)
	}

	return true
}

func (vm *VM) ExecuteContext(gctx context.Context, ctx value.ExecuteContext) (value.Value, error) {
	if len(vm.reg) < vm.top+ctx.Code.NumRegs {
		panic("out of registers")
	}

	var (
		ip   int
		sp   = vm.top
		reg  = vm.reg[sp:]
		seq  = ctx.Code.Instructions
		lits = ctx.Code.Literals
	)

	// Restore the top of the register file
	defer func(v int) { vm.top = v }(vm.top)

	vm.top += ctx.Code.NumRegs

	// TODO use overlapping call args with locals in invoked lambda
	// rather than copy them

	for i, v := range ctx.Args {
		reg[i] = v
	}

	max := len(seq)

	for ip < max {
		i := seq[ip]

		ip++

		/*
			fmt.Printf("@ %d/%p => %s r0:%d r1:%d r2:%d data:%d rest1:%d rest2:%d\n",
				ip, seq,
				i.Op(), i.R0(), i.R1(), i.R2(),
				i.Data(), i.Rest1(), i.Rest2())
		*/
		// fmt.Printf("regs: %d lits: %+v\n", len(reg), lits)

		switch i.Op() {
		case insn.Noop:
			// nothing
		case insn.Reset:
			reg[i.R0()] = nil
		case insn.StoreInt:
			reg[i.R0()] = value.MakeI64(i.Data())
		case insn.CopyReg:
			reg[i.R0()] = reg[i.R1()]
		case insn.Call0:
			res, err := vm.callN(gctx, reg[i.R1()], nil, lits[i.R2()])
			if err != nil {
				return nil, err
			}

			// fmt.Printf("set %d: %+v (%T)\n", i.R0(), res, res)

			reg[i.R0()] = res
		case insn.CallN:
			res, err := vm.callN(
				gctx,
				reg[i.R1()],
				reg[i.R1()+1:int64(i.R1())+i.Rest2()+1],
				lits[i.R2()],
			)
			if err != nil {
				return nil, err
			}

			reg[i.R0()] = res
		case insn.GIF:
			if !vm.isTrue(reg[i.R0()]) {
				ip = int(i.Data())
			}
		case insn.Goto:
			ip = int(i.Data())
		case insn.CreateLambda:
			reg[i.R0()] = vm.createLambda(ctx, i.R1(), vm.refs(ctx, ip, i.R2(), i.R2()), i.Rest2())
			ip += i.R2()
		case insn.Invoke:
			res, err := vm.invoke(gctx, reg[i.R1():i.R1()+int(i.Rest1()+1)])
			if err != nil {
				return nil, err
			}

			reg[i.R0()] = res
		case insn.Return:
			return reg[i.R0()], nil
		case insn.ReadRef:
			reg[i.R0()] = ctx.Refs[i.Data()].Value
		case insn.StoreRef:
			ctx.Refs[i.R0()].Value = reg[i.R1()]
		case insn.GetMirror:
			reg[i.R0()] = vm.getMirror(gctx, reg[i.R0()])
		case insn.Self:
			reg[i.R0()] = ctx.Self
		case insn.String:
			reg[i.R0()] = vm.NewString(ctx.Code.Literals[i.R1()])
		case insn.GetScoped:
			reg[i.R0()] = vm.getScoped(gctx, ctx.Code.Literals[i.R1()])
		case insn.NewList:
			reg[i.R0()] = value.NewList(vm, i.R1())
		case insn.ListAppend:
			reg[i.R0()].(*value.List).Append(reg[i.R1()])
		case insn.SetIvar:
			no := ctx.Self.(*value.NativeObject)
			no.Ivars[no.Class(vm).Ivars[ctx.Code.Literals[i.R1()]]] = reg[i.R0()]
		case insn.GetIvar:
			no := ctx.Self.(*value.NativeObject)
			reg[i.R0()] = no.Ivars[no.Class(vm).Ivars[ctx.Code.Literals[i.R1()]]]
		default:
			panic(fmt.Sprintf("unknown op: %s", i.Op()))
		}
	}

	return nil, nil
}

func (vm *VM) refs(ctx value.ExecuteContext, ip, sz, cap int) []*value.Ref {
	if sz == 0 {
		return nil
	}

	refs := make([]*value.Ref, 0, cap)

	for i := 0; i < sz; i++ {
		c := ctx.Code.Instructions[ip+i]
		if c.Op() != insn.ReadRef {
			panic("expected readref")
		}

		refs = append(refs, ctx.Refs[c.R1()])
	}

	for i := sz; i < cap; i++ {
		refs = append(refs, &value.Ref{})
	}

	return refs
}

func (vm *VM) createLambda(ctx value.ExecuteContext, args int, refs []*value.Ref, code int64) *value.Lambda {
	if int(code) >= len(ctx.Code.SubCode) {
		panic(fmt.Sprintf("Missing code: %d", code))
	}
	return value.CreateLambda(vm, ctx.Code.SubCode[code], ctx.Self, refs, args)
}

func (vm *VM) getMirror(ctx context.Context, obj value.Value) value.Value {
	cls := vm.Registry().Mirror

	val, err := vm.callN(ctx, cls, []value.Value{obj}, "resolve")
	if err != nil {
		panic(err)
	}

	return val
}

func (vm *VM) getScoped(ctx context.Context, name string) value.Value {
	if val, ok := value.GetScoped(ctx, name); ok {
		return val
	}

	return vm.Nil()
}
