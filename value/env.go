package value

type ExecuteContext struct {
	Code *Code
	Refs []*Ref
	Args []Value
}

type Env interface {
	MustFindClass(name string) *Class
	ArgumentError(expected, received int) (Value, error)
	Class() *Class
	I64Class() *Class
	BoolClass() *Class
	LambdaClass() *Class
	ExecuteContext(ExecuteContext) (Value, error)
}
