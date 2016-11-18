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
	NumRegs  int
	Literals []string
	Sequence []insn.Instruction
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
	if len(vm.reg) < ctx.NumRegs {
		vm.reg = make([]value.Value, ctx.NumRegs)
	}

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
			vm.reg[i.R0()] = nil
		case insn.StoreInt:
			vm.reg[i.R0()] = builtin.MakeI64(i.Data())
		case insn.CopyReg:
			vm.reg[i.R0()] = vm.reg[i.R1()]
		case insn.Call0:
			res, err := vm.invokeN(
				vm.reg[i.R1()],
				nil,
				ctx.Literals[i.R2()],
			)
			if err != nil {
				return err
			}

			// fmt.Printf("set %d: %+v (%T)\n", i.R0(), res, res)

			vm.reg[i.R0()] = res
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
			if !isTrue(vm.reg[i.R0()]) {
				ip = int(i.Data())
			}
		case insn.Goto:
			ip = int(i.Data())
		default:
			panic(fmt.Sprintf("unknown op: %s", i.Op()))
		}
	}

	return nil
}
