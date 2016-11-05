package vm

import "github.com/evanphx/m13/insn"

type Value interface {
}

type Integer int64

func (i Integer) Value() int64 {
	return int64(i)
}

type VM struct {
	reg []Value
}

func NewVM() (*VM, error) {
	return &VM{
		reg: make([]Value, 128),
	}, nil
}

func (vm *VM) Reg(i int) Value {
	return vm.reg[i]
}

func (vm *VM) ExecuteSeq(seq []insn.Instruction) error {
	for _, i := range seq {
		switch i.Op() {
		case insn.StoreInt:
			vm.reg[i.R0()] = Integer(i.Data())
		default:
			panic("unknown op")
		}
	}

	return nil
}
