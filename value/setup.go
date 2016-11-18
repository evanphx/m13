package value

type Setup interface {
	OpenPackage(name string) *Package
	MakeType(cfg *TypeConfig) *Type
	MakeMethod(cfg *MethodConfig) *Method
}

type setupFunc func(Setup)

var setupFuncs []setupFunc

func RegisterSetup(f setupFunc) error {
	setupFuncs = append(setupFuncs, f)
	return nil
}

type Registry struct {
	packages PackageRegistry
	types    map[string]*Type
}

func NewRegistry() *Registry {
	r := &Registry{
		packages: PackageRegistry{
			Packages: make(map[string]*Package),
		},
		types: make(map[string]*Type),
	}

	for _, f := range setupFuncs {
		f(r)
	}

	return r
}

func (r *Registry) FindType(globalName string) (*Type, bool) {
	t, ok := r.types[globalName]
	if !ok {
		return nil, false
	}

	return t, true
}

func (r *Registry) OpenPackage(name string) *Package {
	if pkg, ok := r.packages.Packages[name]; ok {
		return pkg
	}

	pkg := &Package{
		Name:  name,
		Types: make(map[string]*Type),
	}

	r.packages.Packages[name] = pkg

	return pkg
}

func (r *Registry) MakeType(cfg *TypeConfig) *Type {
	typ := &Type{
		Package: cfg.Package,
		Name:    cfg.Name,
		Methods: cfg.Methods,
	}

	cfg.Package.Types[cfg.Name] = typ

	r.types[cfg.GlobalName] = typ

	return typ
}

func (r *Registry) MakeMethod(cfg *MethodConfig) *Method {
	return &Method{
		Name: cfg.Name,
		F:    cfg.Func,
	}
}
