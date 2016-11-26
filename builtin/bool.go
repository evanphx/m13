//go:generate m13-glue -T Bool

package builtin

import "github.com/evanphx/m13/value"

// m13
type Bool bool

func (b Bool) Class(env value.Env) *value.Class {
	return env.BoolClass()
}
