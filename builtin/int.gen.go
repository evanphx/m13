package builtin

import "github.com/evanphx/m13/value"

var type_Integer *value.Type

func (_ Integer) Type() *value.Type {
	return type_Integer
}

var type_I64 *value.Type

func (_ I64) Type() *value.Type {
	return type_I64
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

var type_BigInt *value.Type

func (_ BigInt) Type() *value.Type {
	return type_BigInt
}

func init() {
	pkg := value.OpenPackage("builtin")

	var methods map[string]*value.Method

	methods = make(map[string]*value.Method)

	type_Integer = value.MakeType(&value.TypeConfig{
		Package: pkg,
		Name:    "Integer",
		Parent:  "",
		Methods: methods,
	})

	methods = make(map[string]*value.Method)

	methods["+"] = value.MakeMethod(&value.MethodConfig{
		Name: "+",
		Func: add_adapter,
	})

	type_I64 = value.MakeType(&value.TypeConfig{
		Package: pkg,
		Name:    "I64",
		Parent:  "Integer",
		Methods: methods,
	})

	methods = make(map[string]*value.Method)

	type_BigInt = value.MakeType(&value.TypeConfig{
		Package: pkg,
		Name:    "BigInt",
		Parent:  "Integer",
		Methods: methods,
	})

}
