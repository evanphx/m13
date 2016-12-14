package value

// m13
type String struct {
	Object
	String string
}

// m13 name=+
func (s *String) add(o *String) (*String, error) {
	var ret String = *s

	ret.String += o.String

	return &ret, nil
}
