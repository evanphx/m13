package value

type Value interface {
	Class(env Env) *Class
	Hash() uint64
}
