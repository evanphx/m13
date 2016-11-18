package builtin

import "github.com/evanphx/m13/value"

func (_ Bool) Type(env value.Env) *value.Type {
	return env.MustFindType("builtin.Bool")
}

func setup_Bool(setup value.Setup) {
	pkg := setup.OpenPackage("builtin")

	methods := make(map[string]*value.Method)

	setup.MakeType(&value.TypeConfig{
		Package:    pkg,
		Name:       "Bool",
		Parent:     "",
		Methods:    methods,
		GlobalName: "builtin.Bool",
	})
}

var _ = value.RegisterSetup(setup_Bool)

func (_ Integer) Type(env value.Env) *value.Type {
	return env.MustFindType("builtin.Integer")
}

func setup_Integer(setup value.Setup) {
	pkg := setup.OpenPackage("builtin")

	methods := make(map[string]*value.Method)

	setup.MakeType(&value.TypeConfig{
		Package:    pkg,
		Name:       "Integer",
		Parent:     "",
		Methods:    methods,
		GlobalName: "builtin.Integer",
	})
}

var _ = value.RegisterSetup(setup_Integer)

func (_ I64) Type(env value.Env) *value.Type {
	return env.MustFindType("builtin.I64")
}

func add_adapter(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return env.ArgumentError(len(args), 1)
	}

	self := recv.(I64)

	a0 := args[0].(I64)

	ret, err := self.add(

		a0,
	)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func inc_adapter(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
	if len(args) != 0 {
		return env.ArgumentError(len(args), 0)
	}

	self := recv.(I64)

	ret, err := self.inc()

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func lt_adapter(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return env.ArgumentError(len(args), 1)
	}

	self := recv.(I64)

	a0 := args[0].(I64)

	ret, err := self.lt(

		a0,
	)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func setup_I64(setup value.Setup) {
	pkg := setup.OpenPackage("builtin")

	methods := make(map[string]*value.Method)

	methods["+"] = setup.MakeMethod(&value.MethodConfig{
		Name: "+",
		Func: add_adapter,
	})

	methods["++"] = setup.MakeMethod(&value.MethodConfig{
		Name: "++",
		Func: inc_adapter,
	})

	methods["<"] = setup.MakeMethod(&value.MethodConfig{
		Name: "<",
		Func: lt_adapter,
	})

	setup.MakeType(&value.TypeConfig{
		Package:    pkg,
		Name:       "I64",
		Parent:     "Integer",
		Methods:    methods,
		GlobalName: "builtin.I64",
	})
}

var _ = value.RegisterSetup(setup_I64)

func (_ BigInt) Type(env value.Env) *value.Type {
	return env.MustFindType("builtin.BigInt")
}

func setup_BigInt(setup value.Setup) {
	pkg := setup.OpenPackage("builtin")

	methods := make(map[string]*value.Method)

	setup.MakeType(&value.TypeConfig{
		Package:    pkg,
		Name:       "BigInt",
		Parent:     "Integer",
		Methods:    methods,
		GlobalName: "builtin.BigInt",
	})
}

var _ = value.RegisterSetup(setup_BigInt)
