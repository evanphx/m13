package value

type PackageRegistry struct {
	Packages map[string]*Package
}

var Registry = &PackageRegistry{
	Packages: make(map[string]*Package),
}

type Package struct {
	Types map[string]*Type
}

func OpenPackage(name string) *Package {
	if pkg, ok := Registry.Packages[name]; ok {
		return pkg
	}

	pkg := &Package{
		Types: make(map[string]*Type),
	}

	Registry.Packages[name] = pkg

	return pkg
}

type TypeConfig struct {
	Package *Package
	Name    string
	Parent  string
	Methods map[string]*Method
}

type Type struct {
	Package *Package
	Name    string
	Parent  *Type
	Methods map[string]*Method
}

func MakeType(cfg *TypeConfig) *Type {
	typ := &Type{
		Package: cfg.Package,
		Name:    cfg.Name,
		Methods: cfg.Methods,
	}

	cfg.Package.Types[cfg.Name] = typ

	return typ
}

type Env interface {
	ArgumentError(expected, received int) (Value, error)
}

type MethodConfig struct {
	Name string
	Func func(env Env, recv Value, args []Value) (Value, error)
}

type Method struct {
	Name string
	F    func(env Env, recv Value, args []Value) (Value, error)
}

func MakeMethod(cfg *MethodConfig) *Method {
	return &Method{
		Name: cfg.Name,
		F:    cfg.Func,
	}
}
