package value

type Env interface {
	MustFindClass(name string) *Class
	ArgumentError(expected, received int) (Value, error)
	Class() *Class
	I64Class() *Class
	BoolClass() *Class
}
