package insn

type Op int8

const (
	Noop     Op = 0
	StoreInt Op = 1
)

type Instruction int64

const (
	OpMask    = 0xFF
	Reg0Shift = 8
	Reg0Mask  = 0xFF
	DataShift = 16
)

func (i Instruction) Op() Op {
	return Op(i & OpMask)
}

func (i Instruction) R0() int {
	return int((i >> Reg0Shift) & Reg0Mask)
}

func (i Instruction) Data() int64 {
	return int64(i >> DataShift)
}

type Int int

func Store(reg int, i Int) Instruction {
	var out Instruction

	out |= Instruction(StoreInt)
	out |= (Instruction(reg) << Reg0Shift)
	out |= (Instruction(i) << DataShift)

	return out
}
