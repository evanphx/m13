package value

import (
	"fmt"
	"io"

	"github.com/evanphx/m13/insn"
)

type Code struct {
	NumRefs      int
	NumRegs      int
	Instructions []insn.Instruction
	Literals     []string
	SubCode      []*Code
}

func (c *Code) Disassemble(w io.Writer) {
	fmt.Fprintf(w, "Refs: %d Regs: %d\n", c.NumRefs, c.NumRegs)
	for idx, i := range c.Instructions {
		fmt.Fprintf(w, "%03d ", idx)

		switch i.Op() {
		case insn.CopyReg:
			fmt.Fprintf(w, "r%d = r%d\n",
				i.R0(),
				i.R1())
		case insn.StoreInt:
			fmt.Fprintf(w, "r%d = int(%d)\n",
				i.R0(),
				i.R1())
		case insn.GetMirror:
			fmt.Fprintf(w, "r%d = mirror(r%d)\n",
				i.R0(),
				i.R1())
		case insn.GetScoped:
			fmt.Fprintf(w, "r%d = $%s\n",
				i.R0(),
				c.Literals[i.R1()])
		case insn.String:
			fmt.Fprintf(w, "r%d = \"%s\"\n",
				i.R0(),
				c.Literals[i.R1()])
		case insn.CallN:
			switch i.Rest2() {
			case 0:
				fmt.Fprintf(w, "r%d = r%d.`%s`()\n",
					i.R0(),
					i.R1(),
					c.Literals[i.R2()])
			case 1:
				fmt.Fprintf(w, "r%d = r%d.`%s`(r%d)\n",
					i.R0(),
					i.R1(),
					c.Literals[i.R2()],
					i.R1()+1)
			default:
				fmt.Fprintf(w, "r%d = r%d.`%s`(r%d to r%d)\n",
					i.R0(),
					i.R1(),
					c.Literals[i.R2()],
					i.R1()+1,
					i.R1()+int(i.Rest2()))
			}
		case insn.CreateLambda:
			fmt.Fprintf(w, "r%d = lambda(args=%d, refs=%d, sub=%d)\n",
				i.R0(),
				i.R1(),
				i.R2(),
				i.Rest2())
		case insn.Return:
			fmt.Fprintf(w, "ret r%d\n", i.R0())
		default:
			fmt.Fprintf(w, "%s %d %d %d %d\n",
				i.Op().String(),
				i.R0(), i.R1(),
				i.R2(), i.Data(),
			)
		}
	}

	for i, sub := range c.SubCode {
		fmt.Fprintf(w, "\n==== Sub %d =====\n", i)
		sub.Disassemble(w)
	}
}
