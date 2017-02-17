package value

import "context"

type List struct {
	Object

	data []Value
}

func methListappend_op(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
	self := recv.(*List)

	v := args[0].(Value)

	{
		self.data = append(self.data, v)
		return self, nil
	}

	return recv, nil
}

func methListeach(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
	self := recv.(*List)

	lambda := args[0].(*Lambda)

	{
		for _, v := range self.data {
			_, err := env.InvokeLambda(ctx, lambda, []Value{v})
			if err != nil {
				return nil, err
			}
		}

		return self, nil
	}

	return recv, nil
}

func initList(pkg *Package, cls *Class) {

	cls.AddMethod(&MethodDescriptor{
		Name: "<<",
		Signature: Signature{
			Required: 1,
		},
		Func: methListappend_op,
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "each",
		Signature: Signature{
			Required: 1,
		},
		Func: methListeach,
	})

}
