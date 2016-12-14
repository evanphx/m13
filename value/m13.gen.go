package value

var methods_Bool = []*MethodDescriptor{}

func setup_Bool(setup Setup) {
	setup.ApplyMethods("builtin.Bool", methods_Bool)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*Method)



		setup.MakeClass(&ClassConfig{
			Package: pkg,
			Name: "Bool",
			Parent: "",
			Methods: methods,
			GlobalName: "builtin.Bool",
		})
	*/
}

var _ = RegisterSetup(setup_Bool)

var methods_Integer = []*MethodDescriptor{}

func setup_Integer(setup Setup) {
	setup.ApplyMethods("builtin.Integer", methods_Integer)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*Method)



		setup.MakeClass(&ClassConfig{
			Package: pkg,
			Name: "Integer",
			Parent: "",
			Methods: methods,
			GlobalName: "builtin.Integer",
		})
	*/
}

var _ = RegisterSetup(setup_Integer)

func I64_add_adapter(env Env, recv Value, args []Value) (Value, error) {
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

var method_desc_I64_add = &MethodDescriptor{
	Name: "+",
	Aliases: []string{

		"add",
	},
	Signature: Signature{
		Required: 1,
		Args: []string{

			"j",
		},
	},
	Func: I64_add_adapter,
}

func I64_inc_adapter(env Env, recv Value, args []Value) (Value, error) {
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

var method_desc_I64_inc = &MethodDescriptor{
	Name: "++",
	Aliases: []string{

		"inc",
	},
	Signature: Signature{
		Required: 0,
		Args:     []string{},
	},
	Func: I64_inc_adapter,
}

func I64_lt_adapter(env Env, recv Value, args []Value) (Value, error) {
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

var method_desc_I64_lt = &MethodDescriptor{
	Name:    "<",
	Aliases: []string{},
	Signature: Signature{
		Required: 1,
		Args: []string{

			"j",
		},
	},
	Func: I64_lt_adapter,
}

var methods_I64 = []*MethodDescriptor{

	method_desc_I64_add,

	method_desc_I64_inc,

	method_desc_I64_lt,
}

func setup_I64(setup Setup) {
	setup.ApplyMethods("builtin.I64", methods_I64)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*Method)


			methods["+"] = setup.MakeMethod(&MethodConfig{
				Name: "+",
				Func: I64_add_adapter,
			})

			methods["++"] = setup.MakeMethod(&MethodConfig{
				Name: "++",
				Func: I64_inc_adapter,
			})

			methods["<"] = setup.MakeMethod(&MethodConfig{
				Name: "<",
				Func: I64_lt_adapter,
			})


		setup.MakeClass(&ClassConfig{
			Package: pkg,
			Name: "I64",
			Parent: "Integer",
			Methods: methods,
			GlobalName: "builtin.I64",
		})
	*/
}

var _ = RegisterSetup(setup_I64)

var methods_BigInt = []*MethodDescriptor{}

func setup_BigInt(setup Setup) {
	setup.ApplyMethods("builtin.BigInt", methods_BigInt)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*Method)



		setup.MakeClass(&ClassConfig{
			Package: pkg,
			Name: "BigInt",
			Parent: "Integer",
			Methods: methods,
			GlobalName: "builtin.BigInt",
		})
	*/
}

var _ = RegisterSetup(setup_BigInt)

func ObjectMirror_ObjectClass_adapter(env Env, recv Value, args []Value) (Value, error) {
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

var method_desc_ObjectMirror_ObjectClass = &MethodDescriptor{
	Name:    "class",
	Aliases: []string{},
	Signature: Signature{
		Required: 0,
		Args:     []string{},
	},
	Func: ObjectMirror_ObjectClass_adapter,
}

var methods_ObjectMirror = []*MethodDescriptor{

	method_desc_ObjectMirror_ObjectClass,
}

func setup_ObjectMirror(setup Setup) {
	setup.ApplyMethods("builtin.ObjectMirror", methods_ObjectMirror)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*Method)


			methods["class"] = setup.MakeMethod(&MethodConfig{
				Name: "class",
				Func: ObjectMirror_ObjectClass_adapter,
			})


		setup.MakeClass(&ClassConfig{
			Package: pkg,
			Name: "ObjectMirror",
			Parent: "",
			Methods: methods,
			GlobalName: "builtin.ObjectMirror",
		})
	*/
}

var _ = RegisterSetup(setup_ObjectMirror)

var methods_Package = []*MethodDescriptor{}

func setup_Package(setup Setup) {
	setup.ApplyMethods("builtin.Package", methods_Package)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*Method)



		setup.MakeClass(&ClassConfig{
			Package: pkg,
			Name: "Package",
			Parent: "",
			Methods: methods,
			GlobalName: "builtin.Package",
		})
	*/
}

var _ = RegisterSetup(setup_Package)

var methods_PackageMirror = []*MethodDescriptor{}

func setup_PackageMirror(setup Setup) {
	setup.ApplyMethods("builtin.PackageMirror", methods_PackageMirror)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*Method)



		setup.MakeClass(&ClassConfig{
			Package: pkg,
			Name: "PackageMirror",
			Parent: "",
			Methods: methods,
			GlobalName: "builtin.PackageMirror",
		})
	*/
}

var _ = RegisterSetup(setup_PackageMirror)

func String_add_adapter(env Env, recv Value, args []Value) (Value, error) {
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

var method_desc_String_add = &MethodDescriptor{
	Name:    "+",
	Aliases: []string{},
	Signature: Signature{
		Required: 1,
		Args: []string{

			"o",
		},
	},
	Func: String_add_adapter,
}

var methods_String = []*MethodDescriptor{

	method_desc_String_add,
}

func setup_String(setup Setup) {
	setup.ApplyMethods("builtin.String", methods_String)
	/*
		pkg := setup.OpenPackage("builtin")

		methods := make(map[string]*Method)


			methods["+"] = setup.MakeMethod(&MethodConfig{
				Name: "+",
				Func: String_add_adapter,
			})


		setup.MakeClass(&ClassConfig{
			Package: pkg,
			Name: "String",
			Parent: "",
			Methods: methods,
			GlobalName: "builtin.String",
		})
	*/
}

var _ = RegisterSetup(setup_String)
