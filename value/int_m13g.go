package value

import "context"

type I64 int64

func methI64equal(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
	self := recv.(I64)

	o := args[0].(I64)

	{
		if self == o {
			return env.True(), nil
		}

		return env.False(), nil
	}

	return recv, nil
}

func methI64add(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
	self := recv.(I64)

	o := args[0].(I64)

	{
		return self + o, nil
	}

	return recv, nil
}

func methI64less__than(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
	self := recv.(I64)

	o := args[0].(I64)

	{
		if self < o {
			return env.True(), nil
		}

		return env.False(), nil
	}

	return recv, nil
}

func methI64inc(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
	self := recv.(I64)

	{
		return self + 1, nil
	}

	return recv, nil
}

func initI64(pkg *Package, cls *Class) {

	cls.AddMethod(&MethodDescriptor{
		Name: "equal",
		Aliases: []string{
			"==",
		},
		Signature: Signature{
			Required: 1,
		},
		Func: methI64equal,
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "add",
		Aliases: []string{
			"+",
		},
		Signature: Signature{
			Required: 1,
		},
		Func: methI64add,
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "less_than",
		Aliases: []string{
			"<",
		},
		Signature: Signature{
			Required: 1,
		},
		Func: methI64less__than,
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "inc",
		Aliases: []string{
			"++",
		},
		Signature: Signature{
			Required: 0,
		},
		Func: methI64inc,
	})

}
