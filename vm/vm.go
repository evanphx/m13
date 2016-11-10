package vm

import (
	"github.com/evanphx/m13/builtin"
	"github.com/evanphx/m13/insn"
	"github.com/evanphx/m13/value"
)

type VM struct {
	reg []value.Value
}

func NewVM() (*VM, error) {
	return &VM{
		reg: make([]value.Value, 128),
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
			panic("unknown op")
		}
	}

	return nil
}

type ExecuteContext struct {
	NumRegs  int
	Literals []string
	Sequence []insn.Instruction
}

func (vm *VM) ExecuteContext(ctx *ExecuteContext) error {
	if len(vm.reg) < ctx.NumRegs {
		vm.reg = make([]value.Value, ctx.NumRegs)
	}

	max := len(ctx.Sequence)

	var ip int
	for ip < max {
		i := ctx.Sequence[ip]

		ip++

		switch i.Op() {
		case insn.Noop:
			// nothing
		case insn.Reset:
			vm.reg[i.R0()] = nil
		case insn.StoreInt:
			vm.reg[i.R0()] = builtin.MakeI64(i.Data())
		case insn.CopyReg:
			vm.reg[i.R0()] = vm.reg[i.R1()]
		case insn.CallN:
			res, err := vm.invokeN(
				vm.reg[i.R1()],
				vm.reg[i.R1()+1:int64(i.R1())+i.Rest2()+1],
				ctx.Literals[i.R2()],
			)
			if err != nil {
				return err
			}

			vm.reg[i.R0()] = res
		case insn.GIF:
			if vm.reg[i.R0()] == nil {
				ip = int(i.Data())
			}
		default:
			panic("unknown op")
		}
	}

	return nil
}
