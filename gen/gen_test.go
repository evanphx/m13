package gen

import (
	"testing"

	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/insn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestGen(t *testing.T) {
	n := neko.Start(t)

	n.It("generates bytecode to store an int", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		err = g.Generate(&ast.Integer{Value: 1})
		require.NoError(t, err)

		seq := g.Sequence()

		require.Equal(t, 1, len(seq))

		i := seq[0]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(1), i.Data())
	})

	n.It("generates bytecode to store a local", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		g.Reserve(1)

		tree := &ast.Assign{
			Name:  "a",
			Value: &ast.Integer{Value: 47},
		}

		err = g.Generate(tree)
		require.NoError(t, err)

		seq := g.Sequence()

		require.Equal(t, 2, len(seq))

		i := seq[0]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 1, i.R0())
		assert.Equal(t, int64(47), i.Data())

		i = seq[1]

		assert.Equal(t, insn.CopyReg, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 1, i.R1())
	})

	n.It("generates bytecode for an operator", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		tree := &ast.Op{
			Name:  "+",
			Left:  &ast.Integer{Value: 3},
			Right: &ast.Integer{Value: 4},
		}

		err = g.Generate(tree)
		require.NoError(t, err)

		seq := g.Sequence()

		require.Equal(t, 3, len(seq))

		i := seq[0]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(3), i.Data())

		i = seq[1]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 1, i.R0())
		assert.Equal(t, int64(4), i.Data())

		i = seq[2]

		assert.Equal(t, insn.CallN, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 0, i.R1())
		assert.Equal(t, 0, i.R2())
		assert.Equal(t, int64(1), i.Rest2())
		assert.Equal(t, "+", g.literals[0])
	})

	n.It("generates bytecode for an if", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		tree := &ast.If{
			Cond: &ast.Integer{Value: 3},
			Body: &ast.Block{
				Expressions: []ast.Node{
					&ast.Integer{Value: 4},
				},
			},
		}

		err = g.Generate(tree)
		require.NoError(t, err)

		seq := g.Sequence()

		require.Equal(t, 3, len(seq))

		i := seq[0]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(3), i.Data())

		i = seq[1]

		assert.Equal(t, insn.GIF, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(3), i.Data())

		i = seq[2]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(4), i.Data())
	})

	n.It("generates bytecode for an inc of a variable", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		tree := &ast.Inc{
			Receiver: &ast.Variable{Name: "a"},
		}

		g.locals["a"] = 7
		g.sp = 8

		err = g.Generate(tree)
		require.NoError(t, err)

		seq := g.Sequence()

		i := seq[0]

		assert.Equal(t, insn.Call0, i.Op())
		assert.Equal(t, 7, i.R0())
		assert.Equal(t, 7, i.R1())
		assert.Equal(t, 0, i.R2())
		assert.Equal(t, int64(0), i.Rest2())

		assert.Equal(t, "++", g.literals[0])
	})

	n.It("generates bytecode for a dec of a variable", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		tree := &ast.Dec{
			Receiver: &ast.Variable{Name: "a"},
		}

		g.locals["a"] = 7
		g.sp = 8

		err = g.Generate(tree)
		require.NoError(t, err)

		seq := g.Sequence()

		i := seq[0]

		assert.Equal(t, insn.Call0, i.Op())
		assert.Equal(t, 7, i.R0())
		assert.Equal(t, 7, i.R1())
		assert.Equal(t, 0, i.R2())
		assert.Equal(t, int64(0), i.Rest2())

		assert.Equal(t, "--", g.literals[0])
	})

	n.It("generates bytecode for a while", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		tree := &ast.While{
			Cond: &ast.Integer{Value: 3},
			Body: &ast.Block{
				Expressions: []ast.Node{
					&ast.Integer{Value: 4},
				},
			},
		}

		err = g.Generate(tree)
		require.NoError(t, err)

		seq := g.Sequence()

		require.Equal(t, 4, len(seq))

		i := seq[0]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(3), i.Data())

		i = seq[1]

		assert.Equal(t, insn.GIF, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(4), i.Data())

		i = seq[2]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(4), i.Data())

		i = seq[3]

		assert.Equal(t, insn.Goto, i.Op())
		assert.Equal(t, int64(0), i.Data())
	})

	n.It("generates bytecode for a lambda", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		tree := &ast.Lambda{
			Expr: &ast.Integer{Value: 3},
		}

		err = g.Generate(tree)
		require.NoError(t, err)

		seq := g.Sequence()

		i := seq[0]

		assert.Equal(t, insn.CreateLambda, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 0, i.R1())
		assert.Equal(t, int64(0), i.Rest1())

		sub := g.subSequences[0]

		i = sub[0]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(3), i.Data())
	})

	n.Meow()
}
