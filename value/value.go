package value

type Value interface {
	Type(env Env) *Type
}
