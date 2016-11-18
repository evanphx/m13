package value

type Env interface {
	MustFindType(name string) *Type
	ArgumentError(expected, received int) (Value, error)
}
