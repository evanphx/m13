package builtin

import "github.com/evanphx/m13/value"

// m13
type String struct {
	value.Object
	str string
}

// m13 name=+
func (s *String) add(o *String) (*String, error) {
	var ret String = *s

	ret.str += o.str

	return &ret, nil
}
