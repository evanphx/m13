package value

import "context"

type Env interface {
	Nil() Value
	True() Value
	False() Value
	MustFindClass(name string) *Class
	ArgumentError(expected, received int) (Value, error)
	TypeError(value Value, expected string) (Value, error)
	Class() *Class
	ObjectClass() *Class
	I64Class() *Class
	BoolClass() *Class
	LambdaClass() *Class
	StringClass() *Class
	NewString(str string) *String
	ListClass() *Class
	MapClass() *Class
	IOClass() *Class
	ExecuteContext(context.Context, ExecuteContext) (Value, error)
	InvokeLambda(ctx context.Context, lamb *Lambda, args []Value) (Value, error)
}

type ExecuteContext struct {
	Env  Env
	Code *Code
	Refs []*Ref
	Self Value
	Args []Value
}
