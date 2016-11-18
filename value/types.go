package value

import "fmt"

type PackageRegistry struct {
	Packages map[string]*Package
}

type Package struct {
	Name  string
	Types map[string]*Type
}

type TypeConfig struct {
	Package    *Package
	Name       string
	GlobalName string
	Parent     string
	Methods    map[string]*Method
}

type Type struct {
	Package    *Package
	Name       string
	GlobalName string
	Parent     *Type
	Methods    map[string]*Method
}

func (t *Type) FullName() string {
	return fmt.Sprintf("%s.%s", t.Package.Name, t.Name)
}

type MethodConfig struct {
	Name string
	Func func(env Env, recv Value, args []Value) (Value, error)
}

type Method struct {
	Name string
	F    func(env Env, recv Value, args []Value) (Value, error)
}
