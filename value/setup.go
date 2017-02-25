package value

import (
	"fmt"
	"strings"
)

type Setup interface {
	OpenPackage(name string) *Package
	MakeClass(cfg *ClassConfig) *Class
	MakeMethod(cfg *MethodConfig) *Method
	ApplyMethods(path string, methods []*MethodDescriptor)
}

type setupFunc func(Setup)

var setupFuncs []setupFunc

func RegisterSetup(f setupFunc) error {
	setupFuncs = append(setupFuncs, f)
	return nil
}

type Registry struct {
	packages PackageRegistry
	types    map[string]*Class

	NilClass  *Class
	Object    *Class
	Class     *Class
	BoolClass *Class
	I64Class  *Class
	Mirror    *Class
	Package   *Class
	Lambda    *Class
	String    *Class
	List      *Class
	IO        *Class
	Map       *Class
}

func NewRegistry() *Registry {
	r := &Registry{
		packages: PackageRegistry{
			Packages: make(map[string]*Package),
		},
		types: make(map[string]*Class),
	}

	err := r.Boot()
	if err != nil {
		panic(err)
	}

	for _, f := range setupFuncs {
		f(r)
	}

	return r
}

func (r *Registry) FindClass(globalName string) (*Class, bool) {
	t, ok := r.types[globalName]
	if !ok {
		return nil, false
	}

	return t, true
}

func (r *Registry) MustFindClass(name string) *Class {
	cls, ok := r.FindClass(name)
	if !ok {
		panic("Unable to find class")
	}

	return cls
}

func (r *Registry) OpenPackage(name string) *Package {
	if pkg, ok := r.packages.Packages[name]; ok {
		return pkg
	}

	pkg := &Package{
		Name:    name,
		Classes: make(map[string]*Class),
	}

	sing := r.NewClass(pkg, "$pkg", r.Package)

	pkg.SetClass(sing)

	r.packages.Packages[name] = pkg

	return pkg
}

func (r *Registry) MakeClass(cfg *ClassConfig) *Class {
	typ := &Class{
		Package: cfg.Package,
		Name:    cfg.Name,
		Methods: cfg.Methods,
	}

	cfg.Package.Classes[cfg.Name] = typ

	r.types[cfg.GlobalName] = typ

	return typ
}

func (r *Registry) MakeMethod(cfg *MethodConfig) *Method {
	return &Method{
		Name: cfg.Name,
		Func: cfg.Func,
	}
}

func (r *Registry) ResolveClass(path string) (*Class, error) {
	dot := strings.LastIndexByte(path, '.')

	var (
		pkg string
		cls string
	)

	if dot == -1 {
		pkg = "builtin"
		cls = path
	} else {
		pkg = path[:dot]
		cls = path[dot+1:]
	}

	if pkg, ok := r.packages.Packages[pkg]; ok {
		if obj, ok := pkg.Classes[cls]; ok {
			return obj, nil
		} else {
			return nil, fmt.Errorf("Unable to resolve class: %s", path)
		}
	} else {
		return nil, fmt.Errorf("Unable to resolve package: %s", pkg)
	}
}

func (r *Registry) ApplyMethods(name string, methods []*MethodDescriptor) {
	cls, err := r.ResolveClass(name)
	if err != nil {
		panic(err)
	}

	for _, desc := range methods {
		cls.AddMethod(desc)
	}
}
