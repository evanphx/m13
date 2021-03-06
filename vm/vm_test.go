package vm

import (
	"context"
	"testing"

	"github.com/evanphx/m13/insn"
	"github.com/evanphx/m13/value"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestVM(t *testing.T) {
	n := neko.Start(t)

	b := insn.Builder

	n.It("can store an integer into a register", func() {
		var seq []insn.Instruction

		seq = append(seq, b.Store(0, insn.Int(1)))

		vm, err := NewVM()
		require.NoError(t, err)

		ctx := value.ExecuteContext{
			Code: &value.Code{
				NumRegs:      1,
				Instructions: seq,
			},
		}

		_, err = vm.ExecuteContext(context.TODO(), ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(value.I64)
		require.True(t, ok)

		assert.Equal(t, value.I64(1), val)
	})

	n.It("can copy a value from one register to another", func() {
		var seq []insn.Instruction

		seq = append(seq, b.StoreReg(0, 1))

		vm, err := NewVM()
		require.NoError(t, err)

		vm.reg[1] = value.I64(47)

		ctx := value.ExecuteContext{
			Code: &value.Code{
				NumRegs:      1,
				Instructions: seq,
			},
		}

		_, err = vm.ExecuteContext(context.TODO(), ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(value.I64)
		require.True(t, ok)

		assert.Equal(t, value.I64(47), val)
	})

	n.It("can invoke an operator on an integer", func() {
		vm, err := NewVM()
		require.NoError(t, err)

		val, err := vm.callN(context.TODO(), value.I64(3), []value.Value{value.I64(4)}, &value.CallSite{Name: "+"})
		require.NoError(t, err)

		i, ok := val.(value.I64)
		require.True(t, ok)

		assert.Equal(t, value.I64(7), i)
	})

	n.It("calls a method", func() {
		var seq []insn.Instruction

		seq = append(seq, b.Store(0, insn.Int(3)))
		seq = append(seq, b.Store(1, insn.Int(4)))
		seq = append(seq, b.CallOp(0, 0, 0))

		ctx := value.ExecuteContext{
			Code: &value.Code{
				NumRegs:      2,
				Calls:        []*value.CallSite{{Name: "+"}},
				Instructions: seq,
			},
		}

		vm, err := NewVM()
		require.NoError(t, err)

		vm.reg[1] = value.I64(47)

		_, err = vm.ExecuteContext(context.TODO(), ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(value.I64)
		require.True(t, ok)

		assert.Equal(t, value.I64(7), val)
	})

	n.It("jumps over a condition body", func() {
		var seq []insn.Instruction

		seq = append(seq, b.Store(0, insn.Int(3)))
		seq = append(seq, b.StoreNil(1))
		seq = append(seq, b.GotoIfFalse(1, 4))
		seq = append(seq, b.Store(0, insn.Int(4)))
		seq = append(seq, b.Noop())

		ctx := value.ExecuteContext{
			Code: &value.Code{
				NumRegs:      2,
				Instructions: seq,
			},
		}

		vm, err := NewVM()
		require.NoError(t, err)

		_, err = vm.ExecuteContext(context.TODO(), ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(value.I64)
		require.True(t, ok)

		assert.Equal(t, value.I64(3), val)

	})

	n.It("jumps backwards to handle a while", func() {
		var seq []insn.Instruction

		seq = append(seq,
			b.Store(0, insn.Int(0)),
			b.Store(1, insn.Int(3)),
			b.CallOp(2, 0, 0),
			b.GotoIfFalse(2, 7),
			b.Call0(0, 0, 1),
			b.Goto(2),
			b.Noop(),
		)

		ctx := value.ExecuteContext{
			Code: &value.Code{
				NumRegs:      3,
				Instructions: seq,
				Calls:        []*value.CallSite{{Name: "<"}, {Name: "++"}},
			},
		}

		vm, err := NewVM()
		require.NoError(t, err)

		_, err = vm.ExecuteContext(context.TODO(), ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(value.I64)
		require.True(t, ok)

		assert.Equal(t, value.I64(3), val)
	})

	n.It("can create and invoke lambda", func() {
		var seq []insn.Instruction

		seq = append(seq,
			b.CreateLambda(0, 0, 0, 0),
			b.Invoke(0, 0, 0),
			b.Return(0),
		)

		c1 := &value.Code{
			NumRegs: 1,
			Instructions: []insn.Instruction{
				b.Store(0, insn.Int(3)),
				b.Return(0),
			},
		}

		ctx := value.ExecuteContext{
			Code: &value.Code{
				NumRegs:      1,
				Instructions: seq,
				SubCode:      []*value.Code{c1},
			},
		}

		vm, err := NewVM()
		require.NoError(t, err)

		val, err := vm.ExecuteContext(context.TODO(), ctx)
		require.NoError(t, err)

		assert.Equal(t, value.I64(3), val)
	})

	n.It("passes arguments into a lambda", func() {
		var seq []insn.Instruction

		seq = append(seq,
			b.CreateLambda(0, 0, 0, 0),
			b.Store(1, insn.Int(3)),
			b.Invoke(0, 0, 1),
		)

		c1 := &value.Code{
			NumRegs: 1,
			Instructions: []insn.Instruction{
				b.Call0(0, 0, 0),
				b.Return(0),
			},
			Calls: []*value.CallSite{{Name: "++"}},
		}

		ctx := value.ExecuteContext{
			Code: &value.Code{
				NumRegs:      1,
				Instructions: seq,
				SubCode:      []*value.Code{c1},
			},
		}

		vm, err := NewVM()
		require.NoError(t, err)

		_, err = vm.ExecuteContext(context.TODO(), ctx)
		require.NoError(t, err)

		val, ok := vm.reg[0].(value.I64)
		require.True(t, ok)

		assert.Equal(t, value.I64(4), val)
	})

	n.It("can pass and access a ref", func() {
		var seq []insn.Instruction

		seq = append(seq,
			b.Store(0, insn.Int(3)),
			b.StoreRef(0, 0),
			b.CreateLambda(0, 0, 1, 0),
			b.ReadRef(0, 0),
			b.Invoke(0, 0, 0),
			b.Return(0),
		)

		c1 := &value.Code{
			NumRefs: 1,
			NumRegs: 1,
			Instructions: []insn.Instruction{
				b.ReadRef(0, 0),
				b.Return(0),
			},
			Calls: []*value.CallSite{{Name: "++"}},
		}

		ctx := value.ExecuteContext{
			Code: &value.Code{
				NumRefs:      1,
				NumRegs:      1,
				Instructions: seq,
				SubCode:      []*value.Code{c1},
			},
			Refs: make([]*value.Ref, 1),
		}

		ctx.Refs[0] = &value.Ref{}

		vm, err := NewVM()
		require.NoError(t, err)

		out, err := vm.ExecuteContext(context.TODO(), ctx)
		require.NoError(t, err)

		val, ok := out.(value.I64)
		require.True(t, ok)

		assert.Equal(t, value.I64(3), val)
	})

	n.Meow()
}
