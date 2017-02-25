package value

// m13
type Bool bool

func (b Bool) Class(env Env) *Class {
	return env.BoolClass()
}

func (b Bool) Hash() uint64 {
	if b {
		return 1
	} else {
		return 0
	}
}
