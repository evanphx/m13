package gen

import (
	"fmt"

	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/insn"
)

type Generator struct {
	seq []insn.Instruction

	sp int

	literals []string
}

func NewGenerator() (*Generator, error) {
	return &Generator{}, nil
}

func (g *Generator) Reserve(slot int) {
	g.sp = slot
}

func (g *Generator) Sequence() []insn.Instruction {
	return g.seq
}

func (g *Generator) Generate(gn ast.Node) error {
	switch n := gn.(type) {
	case *ast.Integer:
		g.seq = append(g.seq, insn.Store(g.sp, insn.Int(n.Value)))
	case *ast.Assign:
		err := g.Generate(n.Value)
		if err != nil {
			return err
		}

		g.seq = append(g.seq, insn.StoreReg(0, g.sp))
	case *ast.Op:
		err := g.Generate(n.Left)
		if err != nil {
			return err
		}

		g.sp++

		err = g.Generate(n.Right)
		if err != nil {
			return err
		}

		g.sp--

		idx := len(g.literals)
		g.literals = append(g.literals, n.Name)

		g.seq = append(g.seq, insn.CallOp(g.sp, g.sp, idx))
	default:
		return fmt.Errorf("Unhandled ast type: %T", gn)
	}

	return nil
}
