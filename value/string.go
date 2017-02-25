package value

import "context"
import "hash/fnv"

// m13
type String struct {
	Object
	String string
}

func (s *String) Hash() uint64 {
	h := fnv.New64()
	h.Write([]byte(s.String))
	return h.Sum64()
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
			s2, ok := args[0].(*String)
			if !ok {
				return env.False(), nil
			}

			if s1.String == s2.String {
				return env.True(), nil
			}

			return env.False(), nil
		},
	})
}
