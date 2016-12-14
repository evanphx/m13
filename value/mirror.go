package value

// m13
type ObjectMirror struct {
	Object
	Val Value
}

// m13 name=class
func (m *ObjectMirror) ObjectClass(env Env) (*Class, error) {
	return TrueClass(env, m.Val), nil
}
