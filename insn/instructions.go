package insn

//go:generate stringer -type=Op
type Op int8

const (
	Noop         Op = 0
	StoreInt     Op = 1
	CopyReg      Op = 2
	CallN        Op = 3
	Reset        Op = 4
	Return       Op = 5
	GIF          Op = 6
	Call0        Op = 7
	Goto         Op = 8
	CreateLambda Op = 9
	Invoke       Op = 10
)

type Instruction int64

const (
	OpMask    = 0xFF
	Reg0Shift = 8
	Reg0Mask  = 0xFF
	DataShift = 16

	Reg1Shift  = 16
	Reg1Mask   = 0xFF
	Rest1Shift = 32

	Reg2Shift  = 24
	Reg2Mask   = 0xFF
	Rest2Shift = 32
)

func (i Instruction) Op() Op {
	return Op(i & OpMask)
}

func (i Instruction) R0() int {
	return int((i >> Reg0Shift) & Reg0Mask)
}

func (i Instruction) R1() int {
	return int((i >> Reg1Shift) & Reg1Mask)
}

func (i Instruction) R2() int {
	return int((i >> Reg2Shift) & Reg2Mask)
}

func (i Instruction) Data() int64 {
	return int64(i >> DataShift)
}

func (i Instruction) Rest1() int64 {
	return int64(i >> Rest1Shift)
}

func (i Instruction) Rest2() int64 {
	return int64(i >> Rest2Shift)
}

type BuilderType struct{}

type Int int

func (_ BuilderType) Store(reg int, i Int) Instruction {
	var out Instruction

	out |= Instruction(StoreInt)
	out |= (Instruction(reg) << Reg0Shift)
	out |= (Instruction(i) << DataShift)

	return out
}

func (_ BuilderType) StoreReg(dest, src int) Instruction {
	var out Instruction

	out |= Instruction(CopyReg)
	out |= (Instruction(dest) << Reg0Shift)
	out |= (Instruction(src) << Reg1Shift)

	return out
}

func (_ BuilderType) StoreNil(dest int) Instruction {
	var out Instruction

	out |= Instruction(Reset)
	out |= (Instruction(dest) << Reg0Shift)

	return out
}

func (_ BuilderType) CallOp(dest, base, lit int) Instruction {
	var out Instruction

	out |= Instruction(CallN)
	out |= (Instruction(dest) << Reg0Shift)
	out |= (Instruction(base) << Reg1Shift)
	out |= (Instruction(lit) << Reg2Shift)
	out |= (Instruction(1) << Rest2Shift)

	return out
}

func (_ BuilderType) GotoIfFalse(reg int, pos int) Instruction {
	var out Instruction

	out |= Instruction(GIF)
	out |= (Instruction(reg) << Reg0Shift)
	out |= (Instruction(pos) << DataShift)

	return out
}

func (_ BuilderType) Call0(dest, recv, lit int) Instruction {
	var out Instruction

	out |= Instruction(Call0)
	out |= (Instruction(dest) << Reg0Shift)
	out |= (Instruction(recv) << Reg1Shift)
	out |= (Instruction(lit) << Reg2Shift)

	return out
}

func (_ BuilderType) Goto(dest int) Instruction {
	var out Instruction

	out |= Instruction(Goto)
	out |= (Instruction(dest) << DataShift)

	return out
}

func (_ BuilderType) Noop() Instruction {
	return Instruction(Noop)
}

func (_ BuilderType) CreateLambda(dest, args, sub int) Instruction {
	var out Instruction

	out |= Instruction(CreateLambda)
	out |= (Instruction(dest) << Reg0Shift)
	out |= (Instruction(args) << Reg1Shift)
	out |= (Instruction(sub) << Rest1Shift)

	return out
}

func (_ BuilderType) Invoke(dest, obj, args int) Instruction {
	var out Instruction

	out |= Instruction(Invoke)
	out |= (Instruction(dest) << Reg0Shift)
	out |= (Instruction(obj) << Reg1Shift)
	out |= (Instruction(args) << Rest1Shift)

	return out
}

var Builder BuilderType
