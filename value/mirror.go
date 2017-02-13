package value

import "context"

// m13
type ObjectMirror struct {
	Object
	Val Value
}

// m13 name=class
func (m *ObjectMirror) ObjectClass(env Env) (*Class, error) {
	return TrueClass(env, m.Val), nil
}

func initObjectMirror(cls *Class) {
	cc := &CondDispatcher{}

	cls.AddClassMethod(&MethodDescriptor{
		Name: "resolve",
		Signature: Signature{
			Required: 1,
		},
		Object: cc,
		Func:   cc.Exec,
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "inspect",
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			om := recv.(*ObjectMirror)
			return env.NewString(Inspect(env, om.Val)), nil
		},
	})
}

func initFinalObjectMirror(cls *Class) {
	cls.AddClassMethodCase("resolve", CondAlways{}, &MethodDescriptor{
		Name: "resolve",
		Signature: Signature{
			Required: 1,
		},
		Func: func(ctx context.Context, env Env, value Value, args []Value) (Value, error) {
			mir := &ObjectMirror{Val: args[0]}
			mir.SetClass(cls)

			return mir, nil
		},
	})

}
