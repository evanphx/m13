package vm

import (
	"fmt"

	"github.com/evanphx/m13/builtin"
	"github.com/evanphx/m13/insn"
	"github.com/evanphx/m13/value"
)

type VM struct {
	registry *value.Registry
	reg      []value.Value
}

func NewVM() (*VM, error) {
	return &VM{
		registry: value.NewRegistry(),
		reg:      make([]value.Value, 128),
	}, nil
}

func (vm *VM) Reg(i int) value.Value {
	return vm.reg[i]
}

type ExecuteContext struct {
	Sp   int
	Code *value.Code
	Refs []*value.Ref
}

func isTrue(v value.Value) bool {
	if v == nil {
		return false
	}

	if b, ok := v.(builtin.Bool); ok {
		return bool(b)
	}

	return true
}

func (vm *VM) ExecuteContext(ctx ExecuteContext) (value.Value, error) {
	if len(vm.reg) < ctx.Sp+ctx.Code.NumRegs {
		panic("out of registers")
	}

	var (
		ip   int
		reg  = vm.reg[ctx.Sp:]
		seq  = ctx.Code.Instructions
		lits = ctx.Code.Literals
	)

	max := len(seq)

	for ip < max {
		i := seq[ip]

		ip++

		// fmt.Printf("=> %s r0:%d r1:%d r2:%d data:%d rest1:%d rest2:%d\n",
		// i.Op(), i.R0(), i.R1(), i.R2(),
		// i.Data(), i.Rest1(), i.Rest2())

		// fmt.Printf("regs: %d lits: %+v\n", len(reg), lits)

		switch i.Op() {
		case insn.Noop:
			// nothing
		case insn.Reset:
			reg[i.R0()] = nil
		case insn.StoreInt:
			reg[i.R0()] = builtin.MakeI64(i.Data())
		case insn.CopyReg:
			reg[i.R0()] = reg[i.R1()]
		case insn.Call0:
			res, err := vm.callN(reg[i.R1()], nil, lits[i.R2()])
			if err != nil {
				return nil, err
			}

			// fmt.Printf("set %d: %+v (%T)\n", i.R0(), res, res)

			reg[i.R0()] = res
		case insn.CallN:
			res, err := vm.callN(
				reg[i.R1()],
				reg[i.R1()+1:int64(i.R1())+i.Rest2()+1],
				lits[i.R2()],
			)
			if err != nil {
				return nil, err
			}

			reg[i.R0()] = res
		case insn.GIF:
			if !isTrue(reg[i.R0()]) {
				ip = int(i.Data())
			}
		case insn.Goto:
			ip = int(i.Data())
		case insn.CreateLambda:
			reg[i.R0()] = vm.createLambda(ctx, i.R1(), vm.refs(ctx, ip, i.R2()), i.Rest2())
			ip += i.R2()
		case insn.Invoke:
			res, err := vm.invoke(ctx, reg[i.R1():i.R1()+int(i.Rest1()+1)])
			if err != nil {
				return nil, err
			}

			reg[i.R0()] = res
		case insn.Return:
			return reg[i.R0()], nil
		case insn.ReadRef:
			reg[i.R0()] = ctx.Refs[i.Data()].Value
		case insn.StoreRef:
			ctx.Refs[i.Data()].Value = reg[i.R0()]
		default:
			panic(fmt.Sprintf("unknown op: %s", i.Op()))
		}
	}

	return nil, nil
}

func (vm *VM) refs(ctx ExecuteContext, ip int, sz int) []*value.Ref {
	if sz == 0 {
		return nil
	}

	refs := make([]*value.Ref, 0, sz)

	for i := 0; i < sz; i++ {
		c := ctx.Code.Instructions[ip+i]
		if c.Op() != insn.ReadRef {
			panic("expected readref")
		}

		refs = append(refs, ctx.Refs[c.R1()])
	}

	return refs
}

func (vm *VM) createLambda(ctx ExecuteContext, args int, refs []*value.Ref, code int64) *value.Lambda {
	if int(code) >= len(ctx.Code.SubCode) {
		panic(fmt.Sprintf("Missing code: %d", code))
	}
	return value.CreateLambda(ctx.Code.SubCode[code], refs, args)
}
