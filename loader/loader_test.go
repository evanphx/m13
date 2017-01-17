package loader

import (
	"testing"

	"github.com/evanphx/m13/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestLoader(t *testing.T) {
	n := neko.Start(t)

	n.It("loads a file for a package", func() {
		lpkg, err := Load("./test")
		require.NoError(t, err)

		assert.Equal(t, "add", lpkg.Methods()[0].Name)
	})

	n.It("generates a Package object", func() {
		lpkg, err := Load("./test")
		require.NoError(t, err)

		v, err := vm.NewVM()
		require.NoError(t, err)

		pkg, err := lpkg.Exec(v, v.Registry())
		require.NoError(t, err)

		assert.Equal(t, "add", pkg.Class(v).Methods["add"].Name)
	})

	n.Meow()
}
