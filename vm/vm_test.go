package vm

import (
	"testing"

	"github.com/evanphx/m13/builtin"
	"github.com/evanphx/m13/insn"
	"github.com/evanphx/m13/value"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestVM(t *testing.T) {
	n := neko.Start(t)

	n.It("can store an integer into a register", func() {
		var seq []insn.Instruction

		seq = append(seq, insn.Builder.Store(0, insn.Int(1)))

		vm, err := NewVM()
		require.NoError(t, err)

		ctx := ExecuteContext{
			Code: &value.Code{
				NumRegs:      1,
				Instructions: seq,
			},
		}

		_, err = vm.ExecuteContext(ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(1), val)
	})

	n.It("can copy a value from one register to another", func() {
		var seq []insn.Instruction

		seq = append(seq, insn.Builder.StoreReg(0, 1))

		vm, err := NewVM()
		require.NoError(t, err)

		vm.reg[1] = builtin.I64(47)

		ctx := ExecuteContext{
			Code: &value.Code{
				NumRegs:      1,
				Instructions: seq,
			},
		}

		_, err = vm.ExecuteContext(ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(47), val)
	})

	n.It("can invoke an operator on an integer", func() {
		vm, err := NewVM()
		require.NoError(t, err)

		val, err := vm.callN(builtin.I64(3), []value.Value{builtin.I64(4)}, "+")
		require.NoError(t, err)

		i, ok := val.(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(7), i)
	})

	n.It("calls a method", func() {
		var seq []insn.Instruction

		seq = append(seq, insn.Builder.Store(0, insn.Int(3)))
		seq = append(seq, insn.Builder.Store(1, insn.Int(4)))
		seq = append(seq, insn.Builder.CallOp(0, 0, 0))

		ctx := ExecuteContext{
			Code: &value.Code{
				NumRegs:      2,
				Literals:     []string{"+"},
				Instructions: seq,
			},
		}

		vm, err := NewVM()
		require.NoError(t, err)

		vm.reg[1] = builtin.I64(47)

		_, err = vm.ExecuteContext(ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(7), val)
	})

	n.It("jumps over a condition body", func() {
		var seq []insn.Instruction

		seq = append(seq, insn.Builder.Store(0, insn.Int(3)))
		seq = append(seq, insn.Builder.StoreNil(1))
		seq = append(seq, insn.Builder.GotoIfFalse(1, 4))
		seq = append(seq, insn.Builder.Store(0, insn.Int(4)))
		seq = append(seq, insn.Builder.Noop())

		ctx := ExecuteContext{
			Code: &value.Code{
				NumRegs:      2,
				Instructions: seq,
			},
		}

		vm, err := NewVM()
		require.NoError(t, err)

		_, err = vm.ExecuteContext(ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(3), val)

	})

	n.It("jumps backwards to handle a while", func() {
		var seq []insn.Instruction

		seq = append(seq,
			insn.Builder.Store(0, insn.Int(0)),
			insn.Builder.Store(1, insn.Int(3)),
			insn.Builder.CallOp(2, 0, 0),
			insn.Builder.GotoIfFalse(2, 6),
			insn.Builder.Call0(0, 0, 1),
			insn.Builder.Goto(2),
			insn.Builder.Noop(),
		)

		ctx := ExecuteContext{
			Code: &value.Code{
				NumRegs:      3,
				Instructions: seq,
				Literals:     []string{"<", "++"},
			},
		}

		vm, err := NewVM()
		require.NoError(t, err)

		_, err = vm.ExecuteContext(ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(3), val)
	})

	n.It("can create and invoke lambda", func() {
		var seq []insn.Instruction

		seq = append(seq,
			insn.Builder.CreateLambda(0, 0, 0),
			insn.Builder.Invoke(0, 0, 0),
			insn.Builder.Return(0),
		)

		c1 := &value.Code{
			NumRegs: 1,
			Instructions: []insn.Instruction{
				insn.Builder.Store(0, insn.Int(3)),
				insn.Builder.Return(0),
			},
		}

		ctx := ExecuteContext{
			Code: &value.Code{
				NumRegs:      1,
				Instructions: seq,
				SubCode:      []*value.Code{c1},
			},
		}

		vm, err := NewVM()
		require.NoError(t, err)

		val, err := vm.ExecuteContext(ctx)
		require.NoError(t, err)

		assert.Equal(t, builtin.I64(3), val)
	})

	n.It("passes arguments into a lambda", func() {
		var seq []insn.Instruction

		seq = append(seq,
			insn.Builder.CreateLambda(0, 0, 0),
			insn.Builder.Store(1, insn.Int(3)),
			insn.Builder.Invoke(0, 0, 1),
		)

		c1 := &value.Code{
			NumRegs: 1,
			Instructions: []insn.Instruction{
				insn.Builder.Call0(0, 0, 0),
				insn.Builder.Return(0),
			},
			Literals: []string{"++"},
		}

		ctx := ExecuteContext{
			Code: &value.Code{
				NumRegs:      1,
				Instructions: seq,
				SubCode:      []*value.Code{c1},
			},
		}

		vm, err := NewVM()
		require.NoError(t, err)

		_, err = vm.ExecuteContext(ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(4), val)
	})
	n.Meow()
}
