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
