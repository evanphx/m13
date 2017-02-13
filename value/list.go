package value

import "context"

type List struct {
	Object

	slice []Value
}

func NewList(env Env, cap int) *List {
	list := &List{}
	list.SetClass(env.ListClass())

	list.slice = make([]Value, 0, cap)

	return list
}

func (list *List) Append(v Value) {
	list.slice = append(list.slice, v)
}

func initList(pkg *Package, cls *Class) {
	cls.AddMethod(&MethodDescriptor{
		Name:    "append",
		Aliases: []string{"<<"},
		Signature: Signature{
			Required: 1,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			list := recv.(*List)
			list.slice = append(list.slice, args...)
			return list, nil
		},
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "each",
		Signature: Signature{
			Required: 1,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			list := recv.(*List)
			lamb := args[0].(*Lambda)

			for _, v := range list.slice {
				_, err := env.InvokeLambda(ctx, lamb, []Value{v})
				if err != nil {
					return nil, err
				}
			}

			return list, nil
		},
	})
}
