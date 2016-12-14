package value

// m13
type Bool bool

func (b Bool) Class(env Env) *Class {
	return env.BoolClass()
}
