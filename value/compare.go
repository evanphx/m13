package value

func Compare(v1, v2 Value) bool {
	switch s := v1.(type) {
	case I64:
		if x, ok := v2.(I64); ok {
			return s == x
		}
	case *String:
		if x, ok := v2.(*String); ok {
			return s.String == x.String
		}
	}

	return false
}
