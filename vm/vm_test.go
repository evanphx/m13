package vm

import (
	"testing"

	"github.com/evanphx/m13/builtin"
	"github.com/evanphx/m13/insn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestVM(t *testing.T) {
	n := neko.Start(t)

	n.It("can store an integer into a register", func() {
		var seq []insn.Instruction

		seq = append(seq, insn.Store(0, insn.Int(1)))

		vm, err := NewVM()
		require.NoError(t, err)

		err = vm.ExecuteSeq(seq)
		require.NoError(t, err)

		val, ok := vm.reg[0].(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(1), val)
	})

	n.It("can copy a value from one register to another", func() {
		var seq []insn.Instruction

		seq = append(seq, insn.StoreReg(0, 1))

		vm, err := NewVM()
		require.NoError(t, err)

		vm.reg[1] = builtin.I64(47)

		err = vm.ExecuteSeq(seq)
		require.NoError(t, err)

		val, ok := vm.reg[0].(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(47), val)
	})

	n.It("can invoke an operator on an integer", func() {
		vm, err := NewVM()
		require.NoError(t, err)

		val, err := vm.invokeOp(builtin.I64(3), builtin.I64(4), "+")
		require.NoError(t, err)

		i, ok := val.(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(7), i)
	})

	n.It("calls a method", func() {
		var seq []insn.Instruction

		seq = append(seq, insn.Store(0, insn.Int(3)))
		seq = append(seq, insn.Store(1, insn.Int(4)))
		seq = append(seq, insn.CallOp(0, 0, 0))

		ctx := &ExecuteContext{
			NumRegs:  2,
			Literals: []string{"+"},
			Sequence: seq,
		}

		vm, err := NewVM()
		require.NoError(t, err)

		vm.reg[1] = builtin.I64(47)

		err = vm.ExecuteContext(ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(7), val)
	})

	n.It("jumps over a condition body", func() {
		var seq []insn.Instruction

		seq = append(seq, insn.Store(0, insn.Int(3)))
		seq = append(seq, insn.StoreNil(1))
		seq = append(seq, insn.GotoIfFalse(1, 4))
		seq = append(seq, insn.Store(0, insn.Int(4)))
		seq = append(seq, insn.Instruction(insn.Noop))

		ctx := &ExecuteContext{
			NumRegs:  2,
			Sequence: seq,
		}

		vm, err := NewVM()
		require.NoError(t, err)

		err = vm.ExecuteContext(ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(3), val)

	})

	n.It("jumps backwards to handle a while", func() {
		var seq []insn.Instruction

		seq = append(seq,
			insn.Store(0, insn.Int(0)),
			insn.Store(1, insn.Int(3)),
			insn.CallOp(2, 0, 0),
			insn.GotoIfFalse(2, 6),
			insn.Builder.Call0(0, 0, 1),
			insn.Builder.Goto(2),
			insn.Builder.Noop(),
		)

		ctx := &ExecuteContext{
			NumRegs:  3,
			Sequence: seq,
			Literals: []string{"<", "++"},
		}

		vm, err := NewVM()
		require.NoError(t, err)

		err = vm.ExecuteContext(ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(builtin.I64)
		require.True(t, ok)

		assert.Equal(t, builtin.I64(3), val)
	})

	n.Meow()
}
