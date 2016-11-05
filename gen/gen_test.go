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

	n.Meow()
}
