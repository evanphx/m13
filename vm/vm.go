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

		// fmt.Printf("=> %s\n", i.Op())

		switch i.Op() {
		case insn.Noop:
			// nothing
		case insn.Reset:
			reg[i.R0()] = nil
		case insn.StoreInt:
			reg[i.R0()] = builtin.MakeI64(i.Data())
		case insn.CopyReg:
			reg[i.R0()] = vm.reg[i.R1()]
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
			reg[i.R0()] = vm.createLambda(ctx, i.R1(), i.Rest1())
		case insn.Invoke:
			res, err := vm.invoke(ctx, reg[i.R1()], i.Rest1())
			if err != nil {
				return nil, err
			}

			reg[i.R0()] = res
		case insn.Return:
			return reg[i.R0()], nil
		default:
			panic(fmt.Sprintf("unknown op: %s", i.Op()))
		}
	}

	return nil, nil
}

func (vm *VM) createLambda(ctx ExecuteContext, args int, code int64) *value.Lambda {
	return value.CreateLambda(ctx.Code.SubCode[code], args)
}
