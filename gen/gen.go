package gen

import (
	"fmt"

	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/insn"
)

type Generator struct {
	seq []insn.Instruction

	sp int

	locals   map[string]int
	literals []string

	subSequences [][]insn.Instruction
}

func NewGenerator() (*Generator, error) {
	g := &Generator{
		locals: make(map[string]int),
	}

	return g, nil
}

func (g *Generator) findLiteral(l string) int {
	for i, x := range g.literals {
		if x == l {
			return i
		}
	}

	i := len(g.literals)

	g.literals = append(g.literals, l)

	return i
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
	case *ast.Block:
		for _, ex := range n.Expressions {
			err := g.Generate(ex)
			if err != nil {
				return err
			}
		}
	case *ast.If:
		err := g.Generate(n.Cond)
		if err != nil {
			return err
		}

		patchSp := g.sp

		patchPos := len(g.seq)

		g.seq = append(g.seq, insn.GotoIfFalse(patchSp, 0))

		err = g.Generate(n.Body)
		if err != nil {
			return err
		}

		g.seq[patchPos] = insn.GotoIfFalse(patchSp, len(g.seq))
	case *ast.While:
		condPos := len(g.seq)

		err := g.Generate(n.Cond)
		if err != nil {
			return err
		}

		patchSp := g.sp

		patchPos := len(g.seq)

		g.seq = append(g.seq, insn.GotoIfFalse(patchSp, 0))

		err = g.Generate(n.Body)
		if err != nil {
			return err
		}

		g.seq = append(g.seq, insn.Builder.Goto(condPos))

		g.seq[patchPos] = insn.GotoIfFalse(patchSp, len(g.seq))

	case *ast.Inc:
		v, ok := n.Receiver.(*ast.Variable)
		if !ok {
			return fmt.Errorf("Unable to inc type: %T", n.Receiver)
		}

		reg := g.locals[v.Name]

		lit := g.findLiteral("++")

		g.seq = append(g.seq, insn.Builder.Call0(reg, reg, lit))

	case *ast.Dec:
		v, ok := n.Receiver.(*ast.Variable)
		if !ok {
			return fmt.Errorf("Unable to inc type: %T", n.Receiver)
		}

		reg := g.locals[v.Name]

		lit := g.findLiteral("--")

		g.seq = append(g.seq, insn.Builder.Call0(reg, reg, lit))

	case *ast.Lambda:
		sub, err := NewGenerator()
		if err != nil {
			return err
		}

		err = sub.Generate(n.Expr)
		if err != nil {
			return err
		}

		pos := len(g.subSequences)

		g.subSequences = append(g.subSequences, sub.Sequence())

		g.seq = append(g.seq, insn.Builder.CreateLambda(g.sp, 0, pos))

	default:
		return fmt.Errorf("Unhandled ast type: %T", gn)
	}

	return nil
}
