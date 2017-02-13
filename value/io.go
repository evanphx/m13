package value

import (
	"context"
	"os"
)

type IO struct {
	Object

	f *os.File
}

func NewIO(env Env, f *os.File) *IO {
	io := &IO{f: f}
	io.SetClass(env.IOClass())
	return io
}

func initIO(r *Package, cls *Class) {
	cls.AddMethod(&MethodDescriptor{
		Name: "puts",
		Signature: Signature{
			Required: 1,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			io := recv.(*IO)
			str := args[0].(*String)

			_, err := io.f.WriteString(str.String)
			if err != nil {
				return nil, err
			}

			_, err = io.f.WriteString("\n")
			if err != nil {
				return nil, err
			}

			return io, nil
		},
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "print",
		Signature: Signature{
			Required: 1,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			io := recv.(*IO)
			str := args[0].(*String)

			_, err := io.f.WriteString(str.String)
			if err != nil {
				return nil, err
			}

			return io, nil
		},
	})
}
