package vm

import (
	"testing"

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

		val, ok := vm.reg[0].(Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), val.Value())
	})

	n.Meow()
}
