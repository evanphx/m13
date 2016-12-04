package builtin

import "github.com/evanphx/m13/value"

var methods_Bool = []*value.MethodDescriptor{}

func setup_Bool(setup value.Setup) {
	setup.ApplyMethods("builtin.Bool", methods_Bool)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*value.Method)



		setup.MakeClass(&value.ClassConfig{
			Package: pkg,
			Name: "Bool",
			Parent: "",
			Methods: methods,
			GlobalName: "builtin.Bool",
		})
	*/
}

var _ = value.RegisterSetup(setup_Bool)

var methods_Integer = []*value.MethodDescriptor{}

func setup_Integer(setup value.Setup) {
	setup.ApplyMethods("builtin.Integer", methods_Integer)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*value.Method)



		setup.MakeClass(&value.ClassConfig{
			Package: pkg,
			Name: "Integer",
			Parent: "",
			Methods: methods,
			GlobalName: "builtin.Integer",
		})
	*/
}

var _ = value.RegisterSetup(setup_Integer)

func I64_add_adapter(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
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

var method_desc_I64_add = &value.MethodDescriptor{
	Name: "+",
	Aliases: []string{

		"add",
	},
	Signature: value.Signature{
		Required: 1,
		Args: []string{

			"j",
		},
	},
	Func: I64_add_adapter,
}

func I64_inc_adapter(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
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

var method_desc_I64_inc = &value.MethodDescriptor{
	Name: "++",
	Aliases: []string{

		"inc",
	},
	Signature: value.Signature{
		Required: 0,
		Args:     []string{},
	},
	Func: I64_inc_adapter,
}

func I64_lt_adapter(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
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

var method_desc_I64_lt = &value.MethodDescriptor{
	Name:    "<",
	Aliases: []string{},
	Signature: value.Signature{
		Required: 1,
		Args: []string{

			"j",
		},
	},
	Func: I64_lt_adapter,
}

var methods_I64 = []*value.MethodDescriptor{

	method_desc_I64_add,

	method_desc_I64_inc,

	method_desc_I64_lt,
}

func setup_I64(setup value.Setup) {
	setup.ApplyMethods("builtin.I64", methods_I64)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*value.Method)


			methods["+"] = setup.MakeMethod(&value.MethodConfig{
				Name: "+",
				Func: I64_add_adapter,
			})

			methods["++"] = setup.MakeMethod(&value.MethodConfig{
				Name: "++",
				Func: I64_inc_adapter,
			})

			methods["<"] = setup.MakeMethod(&value.MethodConfig{
				Name: "<",
				Func: I64_lt_adapter,
			})


		setup.MakeClass(&value.ClassConfig{
			Package: pkg,
			Name: "I64",
			Parent: "Integer",
			Methods: methods,
			GlobalName: "builtin.I64",
		})
	*/
}

var _ = value.RegisterSetup(setup_I64)

var methods_BigInt = []*value.MethodDescriptor{}

func setup_BigInt(setup value.Setup) {
	setup.ApplyMethods("builtin.BigInt", methods_BigInt)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*value.Method)



		setup.MakeClass(&value.ClassConfig{
			Package: pkg,
			Name: "BigInt",
			Parent: "Integer",
			Methods: methods,
			GlobalName: "builtin.BigInt",
		})
	*/
}

var _ = value.RegisterSetup(setup_BigInt)

func ObjectMirror_ObjectClass_adapter(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
	if len(args) != 0 {
		return env.ArgumentError(len(args), 0)
	}

	self := recv.(*ObjectMirror)

	ret, err := self.ObjectClass(

		env,
	)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

var method_desc_ObjectMirror_ObjectClass = &value.MethodDescriptor{
	Name:    "class",
	Aliases: []string{},
	Signature: value.Signature{
		Required: 0,
		Args:     []string{},
	},
	Func: ObjectMirror_ObjectClass_adapter,
}

var methods_ObjectMirror = []*value.MethodDescriptor{

	method_desc_ObjectMirror_ObjectClass,
}

func setup_ObjectMirror(setup value.Setup) {
	setup.ApplyMethods("builtin.ObjectMirror", methods_ObjectMirror)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*value.Method)


			methods["class"] = setup.MakeMethod(&value.MethodConfig{
				Name: "class",
				Func: ObjectMirror_ObjectClass_adapter,
			})


		setup.MakeClass(&value.ClassConfig{
			Package: pkg,
			Name: "ObjectMirror",
			Parent: "",
			Methods: methods,
			GlobalName: "builtin.ObjectMirror",
		})
	*/
}

var _ = value.RegisterSetup(setup_ObjectMirror)

func String_add_adapter(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return env.ArgumentError(len(args), 1)
	}

	self := recv.(*String)

	a0 := args[0].(*String)

	ret, err := self.add(

		a0,
	)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

var method_desc_String_add = &value.MethodDescriptor{
	Name:    "+",
	Aliases: []string{},
	Signature: value.Signature{
		Required: 1,
		Args: []string{

			"o",
		},
	},
	Func: String_add_adapter,
}

var methods_String = []*value.MethodDescriptor{

	method_desc_String_add,
}

func setup_String(setup value.Setup) {
	setup.ApplyMethods("builtin.String", methods_String)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*value.Method)


			methods["+"] = setup.MakeMethod(&value.MethodConfig{
				Name: "+",
				Func: String_add_adapter,
			})


		setup.MakeClass(&value.ClassConfig{
			Package: pkg,
			Name: "String",
			Parent: "",
			Methods: methods,
			GlobalName: "builtin.String",
		})
	*/
}

var _ = value.RegisterSetup(setup_String)
