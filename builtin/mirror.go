package builtin

import "github.com/evanphx/m13/value"

// m13
type ObjectMirror struct {
	value.Object
	Val value.Value
}

// m13 name=class
func (m *ObjectMirror) ObjectClass(env value.Env) (*value.Class, error) {
	return value.TrueClass(env, m.Val), nil
}
