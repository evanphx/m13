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

	n.It("generates bytecode for a call", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		tree := &ast.Call{
			MethodName: "+",
			Receiver:   &ast.Integer{Value: 3},
			Args:       []ast.Node{&ast.Integer{Value: 4}},
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

		sub := g.subSequences[0].Sequence()

		i = sub[0]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(3), i.Data())
	})

	n.It("generates bytecode for a lambda with args", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		tree := &ast.Lambda{
			Args: []string{"a", "b"},
			Expr: &ast.Op{
				Name: "+",
				Left: &ast.Op{
					Name:  "-",
					Left:  &ast.Variable{Name: "a"},
					Right: &ast.Integer{Value: 3},
				},
				Right: &ast.Variable{Name: "b"},
			},
		}

		err = g.Generate(tree)
		require.NoError(t, err)

		seq := g.Sequence()

		i := seq[0]

		assert.Equal(t, insn.CreateLambda, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 2, i.R1())
		assert.Equal(t, 0, i.R2())
		assert.Equal(t, int64(0), i.Rest2())

		sub := g.subSequences[0].Sequence()

		i = sub[0]

		assert.Equal(t, insn.CopyReg, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, 0, i.R1())

		i = sub[1]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 3, i.R0())
		assert.Equal(t, int64(3), i.Data())

		i = sub[2]

		assert.Equal(t, insn.CallN, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, 2, i.R1())
		assert.Equal(t, 0, i.R2())
		assert.Equal(t, int64(1), i.Rest2())

		i = sub[3]

		assert.Equal(t, insn.CopyReg, i.Op())
		assert.Equal(t, 3, i.R0())
		assert.Equal(t, 1, i.R1())

		i = sub[4]

		assert.Equal(t, insn.CallN, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, 2, i.R1())
		assert.Equal(t, 1, i.R2())
		assert.Equal(t, int64(1), i.Rest2())
	})

	n.It("generates bytecode for a lambda with a capture local", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		tree := &ast.Block{
			Expressions: []ast.Node{
				&ast.Assign{
					Name:  "a",
					Value: &ast.Integer{Value: 7},
				},
				&ast.Lambda{
					Expr: &ast.Variable{
						Name: "a",
					},
				},
			},
		}

		err = g.Generate(tree)
		require.NoError(t, err)

		seq := g.Sequence()

		i := seq[0]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(7), i.Data())

		i = seq[1]

		assert.Equal(t, insn.StoreRef, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 0, i.R1())

		i = seq[2]

		assert.Equal(t, insn.CreateLambda, i.Op(), i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 0, i.R1())
		assert.Equal(t, 1, i.R2())
		assert.Equal(t, int64(0), i.Rest2())

		i = seq[3]

		assert.Equal(t, insn.ReadRef, i.Op())
		assert.Equal(t, 0, i.R1())

		sub := g.subSequences[0].Sequence()

		i = sub[0]

		assert.Equal(t, insn.ReadRef, i.Op(), i.Op().String())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 0, i.R1())
	})

	n.It("promotes refs through creating lambdas", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		tree := &ast.Block{
			Expressions: []ast.Node{
				&ast.Assign{
					Name:  "a",
					Value: &ast.Integer{Value: 7},
				},
				&ast.Assign{
					Name:  "b",
					Value: &ast.Integer{Value: 7},
				},
				&ast.Lambda{
					Expr: &ast.Block{
						Expressions: []ast.Node{
							&ast.Variable{
								Name: "a",
							},
							&ast.Lambda{
								Expr: &ast.Variable{
									Name: "b",
								},
							},
						},
					},
				},
			},
		}

		err = g.Generate(tree)
		require.NoError(t, err)

		seq := g.Sequence()

		i := seq[0]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(7), i.Data())

		i = seq[1]

		assert.Equal(t, insn.StoreRef, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 0, i.R1())

		i = seq[2]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, int64(7), i.Data())

		i = seq[3]

		assert.Equal(t, insn.StoreRef, i.Op())
		assert.Equal(t, 1, i.R0())
		assert.Equal(t, 0, i.R1())

		i = seq[4]

		assert.Equal(t, insn.CreateLambda, i.Op(), i.Op().String())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 0, i.R1())
		assert.Equal(t, 2, i.R2())
		assert.Equal(t, int64(0), i.Rest2())

		i = seq[5]

		assert.Equal(t, insn.ReadRef, i.Op())
		assert.Equal(t, 0, i.R1())

		i = seq[6]

		assert.Equal(t, insn.ReadRef, i.Op())
		assert.Equal(t, 1, i.R1())

		subG := g.subSequences[0]
		sub := subG.Sequence()

		i = sub[0]

		assert.Equal(t, insn.ReadRef, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 0, i.R1())

		i = sub[1]

		assert.Equal(t, insn.CreateLambda, i.Op(), i.Op().String())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 0, i.R1())
		assert.Equal(t, 1, i.R2())
		assert.Equal(t, int64(0), i.Rest2())

		i = sub[2]

		assert.Equal(t, insn.ReadRef, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 1, i.R1())

		subG = subG.subSequences[0]
		sub = subG.Sequence()

		i = sub[0]

		assert.Equal(t, insn.ReadRef, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 0, i.R1())
	})

	n.It("uses refs for only captured locals", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		var (
			aa = &ast.Assign{
				Name:  "a",
				Value: &ast.Integer{Value: 7},
			}
			ab = &ast.Assign{
				Name:  "b",
				Value: &ast.Integer{Value: 8},
			}
			ac = &ast.Assign{
				Name:  "c",
				Value: &ast.Integer{Value: 9},
			}
			a1 = &ast.Variable{
				Name: "a",
			}
			b1 = &ast.Variable{
				Name: "b",
			}
			c1 = &ast.Variable{
				Name: "c",
			}
			bl = &ast.Variable{
				Name: "b",
			}
			lam = &ast.Lambda{
				Expr: bl,
			}
			a2 = &ast.Variable{
				Name: "a",
			}
			b2 = &ast.Variable{
				Name: "b",
			}
			c2 = &ast.Variable{
				Name: "c",
			}
		)
		tree := &ast.Block{
			Expressions: []ast.Node{
				aa, ab, ac, a1, b1, c1,
				lam,
				a2, b2, c2,
			},
		}

		err = g.Generate(tree)
		require.NoError(t, err)

		assert.Equal(t, 0, aa.Index)
		assert.False(t, aa.Ref)
		assert.Equal(t, 0, ab.Index)
		assert.True(t, ab.Ref)
		assert.Equal(t, 1, ac.Index)
		assert.False(t, ac.Ref)

		assert.Equal(t, 0, a1.Index)
		assert.False(t, a1.Ref)
		assert.Equal(t, 0, b1.Index)
		assert.True(t, b1.Ref)
		assert.Equal(t, 1, c1.Index)
		assert.False(t, c1.Ref)

		assert.Equal(t, 0, a2.Index)
		assert.False(t, a2.Ref)
		assert.Equal(t, 0, b2.Index)
		assert.True(t, b2.Ref)
		assert.Equal(t, 1, c2.Index)
		assert.False(t, c2.Ref)

		assert.Equal(t, 0, bl.Index)
		assert.True(t, bl.Ref)

		seq := g.Sequence()

		i := seq[0]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, int64(7), i.Data())

		i = seq[1]

		assert.Equal(t, insn.CopyReg, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 2, i.R1())

		i = seq[2]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, int64(8), i.Data())

		i = seq[3]

		assert.Equal(t, insn.StoreRef, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 2, i.R1())

		i = seq[4]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, int64(9), i.Data())

		i = seq[5]

		assert.Equal(t, insn.CopyReg, i.Op())
		assert.Equal(t, 1, i.R0())
		assert.Equal(t, 2, i.R1())

		i = seq[6]

		assert.Equal(t, insn.CopyReg, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, 0, i.R1())

		i = seq[7]

		assert.Equal(t, insn.ReadRef, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, 0, i.R1())

		i = seq[8]

		assert.Equal(t, insn.CopyReg, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, 1, i.R1())

		i = seq[9]

		assert.Equal(t, insn.CreateLambda, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, 0, i.R1())
		assert.Equal(t, 1, i.R2())
		assert.Equal(t, int64(0), i.Rest2())

		i = seq[10]

		assert.Equal(t, insn.ReadRef, i.Op())
		assert.Equal(t, 0, i.R1())

		sub := g.subSequences[0].Sequence()

		i = sub[0]

		assert.Equal(t, insn.ReadRef, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 0, i.R1())

		i = seq[11]

		assert.Equal(t, insn.CopyReg, i.Op(), i.Op().String())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, 0, i.R1())

		i = seq[12]

		assert.Equal(t, insn.ReadRef, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, 0, i.R1())

		i = seq[13]

		assert.Equal(t, insn.CopyReg, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, 1, i.R1())
	})

	n.It("generates bytecode for an invoke", func() {
		g, err := NewGenerator()
		require.NoError(t, err)

		err = g.Generate(&ast.Block{
			Expressions: []ast.Node{
				&ast.Assign{
					Name:  "a",
					Value: &ast.Integer{Value: 0},
				},
				&ast.Invoke{
					Name: "a",
					Args: []ast.Node{
						&ast.Integer{Value: 1},
					},
				},
			},
		})

		require.NoError(t, err)

		seq := g.Sequence()

		i := seq[0]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 1, i.R0())
		assert.Equal(t, int64(0), i.Data())

		i = seq[1]

		assert.Equal(t, insn.CopyReg, i.Op())
		assert.Equal(t, 0, i.R0())
		assert.Equal(t, 1, i.R1())

		i = seq[2]

		assert.Equal(t, insn.CopyReg, i.Op())
		assert.Equal(t, 1, i.R0())
		assert.Equal(t, 0, i.R1())

		i = seq[3]

		assert.Equal(t, insn.StoreInt, i.Op())
		assert.Equal(t, 2, i.R0())
		assert.Equal(t, int64(1), i.Data())

		i = seq[4]

		assert.Equal(t, insn.Invoke, i.Op())
		assert.Equal(t, 1, i.R0())
		assert.Equal(t, 1, i.R1())
		assert.Equal(t, int64(1), i.Rest1())
	})

	n.Meow()
}
