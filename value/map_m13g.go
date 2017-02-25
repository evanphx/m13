package value

import "context"

type Map struct {
	Object

	entries *mapEntries
}

func methMapget(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
	self := recv.(*Map)

	key := args[0].(Value)

	{
		val, ok := self.Get(key)
		if ok {
			return val, nil
		}

		return env.Nil(), nil
	}

	return recv, nil
}

func initMap(pkg *Package, cls *Class) {

	cls.AddMethod(&MethodDescriptor{
		Name: "get",
		Aliases: []string{
			"[]",
		},
		Signature: Signature{
			Required: 1,
		},
		Func: methMapget,
	})

}
