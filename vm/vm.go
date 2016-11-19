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

func (vm *VM) ExecuteSeq(seq []insn.Instruction) error {
	for _, i := range seq {
		switch i.Op() {
		case insn.StoreInt:
			vm.reg[i.R0()] = builtin.MakeI64(i.Data())
		case insn.CopyReg:
			vm.reg[i.R0()] = vm.reg[i.R1()]
		default:
			panic("unknown op - huh?")
		}
	}

	return nil
}

type ExecuteContext struct {
	Sp       int
	NumRegs  int
	Literals []string
	Sequence []insn.Instruction
	SubCode  []*value.Code
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

func (vm *VM) ExecuteContext(ctx *ExecuteContext) error {
	if len(vm.reg) < ctx.Sp+ctx.NumRegs {
		panic("out of registers")
	}

	reg := vm.reg[ctx.Sp:]

	max := len(ctx.Sequence)

	var ip int
	for ip < max {
		i := ctx.Sequence[ip]

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
			res, err := vm.invokeN(
				reg[i.R1()],
				nil,
				ctx.Literals[i.R2()],
			)
			if err != nil {
				return err
			}

			// fmt.Printf("set %d: %+v (%T)\n", i.R0(), res, res)

			reg[i.R0()] = res
		case insn.CallN:
			res, err := vm.invokeN(
				reg[i.R1()],
				reg[i.R1()+1:int64(i.R1())+i.Rest2()+1],
				ctx.Literals[i.R2()],
			)
			if err != nil {
				return err
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
			reg[i.R0()] = vm.invoke(ctx, reg[i.R1()], i.Rest1())
		default:
			panic(fmt.Sprintf("unknown op: %s", i.Op()))
		}
	}

	return nil
}

func (vm *VM) createLambda(ctx *ExecuteContext, args int, code int64) *value.Lambda {
	return value.CreateLambda(ctx.SubCode[code], args)
}

func (vm *VM) invoke(ctx *ExecuteContext, val value.Value, args int64) value.Value {
	l := val.(*value.Lambda)

	sub := &ExecuteContext{
		NumRegs:  l.Code.NumRegs,
		Sequence: l.Code.Instructions,
		Literals: l.Code.Literals,
		Sp:       ctx.Sp + ctx.NumRegs,
	}

	err := vm.ExecuteContext(sub)
	if err != nil {
		panic(err)
	}

	return vm.reg[sub.Sp]
}
