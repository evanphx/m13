package gen

import (
	"fmt"

	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/insn"
)

type Generator struct {
	seq []insn.Instruction
}

func NewGenerator() (*Generator, error) {
	return &Generator{}, nil
}

func (g *Generator) Sequence() []insn.Instruction {
	return g.seq
}

func (g *Generator) Generate(gn ast.Node) error {
	switch n := gn.(type) {
	case *ast.Integer:
		g.seq = append(g.seq, insn.Store(0, insn.Int(n.Value)))
	default:
		return fmt.Errorf("Unhandled ast type: %T", gn)
	}

	return nil
}
