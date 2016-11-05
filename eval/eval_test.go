package eval

import (
	"testing"

	"github.com/evanphx/m13/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestEvaluator(t *testing.T) {
	n := neko.Start(t)

	n.It("can evaluate code", func() {
		ev, err := NewEvaluator()
		require.NoError(t, err)

		val, err := ev.Eval("3")
		require.NoError(t, err)

		i, ok := val.(vm.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(3), i.Value())
	})

	n.Meow()
}
