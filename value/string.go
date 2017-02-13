package value

import "context"

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

func initString(r *Package, cls *Class) {
	cls.AddMethod(&MethodDescriptor{
		Name: "==",
		Signature: Signature{
			Required: 1,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			s1 := recv.(*String)
			s2 := args[0].(*String)

			if s1.String == s2.String {
				return env.True(), nil
			}

			return env.False(), nil
		},
	})
}
