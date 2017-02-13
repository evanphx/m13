package value

import (
	"context"
	"fmt"
)

// m13
type Package struct {
	Object

	Name    string
	Classes map[string]*Class
}

func (p *Package) MustFindClass(name string) *Class {
	cls, ok := p.Classes[name]
	if !ok {
		panic(fmt.Sprintf("Unable to find class: %s", name))
	}

	return cls
}

// m13
type PackageMirror struct {
	Object

	p *Package
}

func (m *PackageMirror) AddMethod(env Env, name *String, l *Lambda) (Value, error) {
	m.p.Class(env).AddMethod(&MethodDescriptor{
		Name: name.String,
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			return env.ExecuteContext(ctx, ExecuteContext{
				Code: l.Code,
				Refs: l.Refs,
				Args: args,
			})
		},
	})

	return nil, nil
}

func initPackageMirror(r *Package, cls *Class) {
	pc := r.MustFindClass("Package")
	mc := r.MustFindClass("ObjectMirror")

	mc.AddClassMethodCase("resolve", NewCheckClass(pc, 0), &MethodDescriptor{
		Name: "resolve",
		Signature: Signature{
			Required: 1,
		},
		Func: func(ctx context.Context, env Env, value Value, args []Value) (Value, error) {
			pm := &PackageMirror{p: args[0].(*Package)}
			pm.SetClass(cls)
			return pm, nil
		},
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "add_method",
		Signature: Signature{
			Required: 2,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			name := args[0].(*String)
			lamb := args[1].(*Lambda)

			return recv.(*PackageMirror).AddMethod(env, name, lamb)
		},
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "add_class",
		Signature: Signature{
			Required: 2,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			pm := recv.(*PackageMirror)
			name := args[0].(*String)
			lamb := args[1].(*Lambda)

			nc := bootClass(pm.p, name.String, env.ObjectClass())
			nc.SetClass(env.Class())

			_, err := env.InvokeLambda(ctx, lamb.RedirectSelf(nc), nil)
			if err != nil {
				return nil, err
			}

			return nc, nil
		},
	})
}
