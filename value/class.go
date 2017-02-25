package value

import (
	"context"
	"fmt"
)

func TrueClass(env Env, v Value) *Class {
	cls := v.Class(env)

	for cls.metaclass {
		cls = cls.Parent
	}

	return cls
}

func ISA(env Env, v Value, target *Class) bool {
	cls := v.Class(env)

	for cls != nil {
		if cls == target {
			return true
		}

		cls = cls.Parent
	}

	return false
}

func (c *Class) LookupMethod(name string) (*Method, bool) {
	for c != nil {
		if t, ok := c.Methods[name]; ok {
			return t, true
		}

		c = c.Parent
	}

	return nil, false
}

func (c *Class) AddMethod(cfg *MethodDescriptor) {
	method := &Method{
		Name:      cfg.Name,
		Signature: cfg.Signature,
		Object:    cfg.Object,
		Func:      cfg.Func,
	}

	c.Methods[cfg.Name] = method

	for _, arg := range cfg.Aliases {
		c.Methods[arg] = method
	}
}

func (c *Class) AliasMethod(from, to string) {
	c.Methods[to] = c.Methods[from]
}

func (c *Class) AddClassMethod(cfg *MethodDescriptor) {
	c.class.AddMethod(cfg)
}

func (c *Class) AddMethodCase(name string, cond Cond, cfg *MethodDescriptor) error {
	var cc *CondDispatcher

	method, ok := c.Methods[name]
	if !ok {
		cc = &CondDispatcher{}

		desc := &Method{
			Name:      cfg.Name,
			Signature: cfg.Signature,
			Object:    cc,
			Func:      cc.Exec,
		}

		c.Methods[name] = desc
	} else {
		cc, ok = method.Object.(*CondDispatcher)
		if !ok {
			return fmt.Errorf("Method is not a CondDispatcher")
		}
	}

	f := CondFunc{
		Cond: cond,
		Method: &Method{
			Name:      cfg.Name,
			Signature: cfg.Signature,
			Object:    cfg.Object,
			Func:      cfg.Func,
		},
	}

	cc.Conditions = append(cc.Conditions, f)
	return nil
}

func (c *Class) AddClassMethodCase(name string, cond Cond, cfg *MethodDescriptor) error {
	return c.class.AddMethodCase(name, cond, cfg)
}

type NativeObject struct {
	Object
	Ivars []Value
}

func initClass(r *Package, cls *Class) {
	cls.AddMethod(&MethodDescriptor{
		Name: "new",
		Signature: Signature{
			Required: -1,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			sc := recv.(*Class)
			obj := &NativeObject{
				Ivars: make([]Value, len(sc.Ivars)),
			}

			obj.SetClass(sc)

			if t, ok := sc.Methods["initialize"]; ok {
				_, err := t.Func(ctx, env, obj, args)
				if err != nil {
					return nil, err
				}
			}

			return obj, nil
		},
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "name",
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			dc := recv.(*Class)

			return env.NewString(dc.GlobalName), nil
		},
	})
}

type ClassMirror struct {
	Object

	cls *Class
}

func initClassMirror(r *Package, cls *Class) {
	pc := r.MustFindClass("Class")
	mc := r.MustFindClass("ObjectMirror")

	mc.AddClassMethodCase("resolve", NewCheckClass(pc, 0), &MethodDescriptor{
		Name: "resolve",
		Signature: Signature{
			Required: 1,
		},
		Func: func(ctx context.Context, env Env, value Value, args []Value) (Value, error) {
			pm := &ClassMirror{cls: args[0].(*Class)}
			pm.SetClass(cls)
			return pm, nil
		},
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "add_ivar",
		Signature: Signature{
			Required: 2,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			name := args[0].(*String)

			dc := recv.(*ClassMirror).cls

			if dc.Ivars == nil {
				dc.Ivars = make(map[string]int)
			}

			dc.Ivars[name.String] = len(dc.Ivars)

			list := args[1].(*List)

			for _, trait := range list.data {
				str := trait.(*String)

				switch str.String {
				case "r":
					dc.AddMethod(&MethodDescriptor{
						Name: name.String,
						Signature: Signature{
							Required: 0,
						},
						Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
							no := recv.(*NativeObject)
							return no.Ivars[no.Class(env).Ivars[name.String]], nil
						},
					})
				}
			}

			return name, nil
		},
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "add_method",
		Signature: Signature{
			Required: 2,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			rc := recv.(*ClassMirror).cls
			name := args[0].(*String)
			lamb := args[1].(*Lambda)

			rc.AddMethod(&MethodDescriptor{
				Name: name.String,
				Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
					return env.InvokeLambda(ctx, lamb.RedirectSelf(recv), args)
				},
			})

			return name, nil
		},
	})

	cls.AddMethod(&MethodDescriptor{
		Name: "alias_method",
		Signature: Signature{
			Required: 2,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			rc := recv.(*ClassMirror).cls
			from := args[0].(*String)
			to := args[1].(*String)

			rc.AliasMethod(from.String, to.String)

			return from, nil
		},
	})
}
