package value

import "fmt"

func bootClass(pkg *Package, name string, parent *Class) *Class {
	cls := &Class{
		Parent:     parent,
		Package:    pkg,
		Name:       name,
		GlobalName: fmt.Sprintf("%s.%s", pkg.Name, name),
		Methods:    map[string]*Method{},
	}

	pkg.Classes[name] = cls

	return cls
}

func (r *Registry) NewClass(pkg *Package, name string, parent *Class) *Class {
	cls := bootClass(pkg, name, parent)
	cls.Object.class = r.Class

	r.types[cls.GlobalName] = cls

	return cls
}

func (c *Class) Metaclass(env Env) *Class {
	if c.class.metaclass {
		return c.class
	}

	mcCls := bootClass(c.Package, "<Class>", nil)
	mcCls.metaclass = true

	if c.class == c {
		c.class = mcCls
		mcCls.class = mcCls
	} else {
		cur := c.class
		c.class = mcCls
		mcCls.class = cur.Metaclass(env)
	}

	super := c.Parent
	for super != nil && super.metaclass {
		super = super.Parent
	}

	if super == nil {
		mcCls.Parent = env.Class()
	} else {
		mcCls.Parent = super.Metaclass(env)
	}

	return mcCls
}

func (r *Registry) Boot() error {
	pkg := r.OpenPackage("builtin")

	obj := bootClass(pkg, "Object", nil)
	cls := bootClass(pkg, "Class", obj)

	obj.class = cls
	cls.class = cls

	r.Object = obj
	r.Class = cls

	r.BoolClass = r.NewClass(pkg, "Bool", obj)
	r.NilClass = r.NewClass(pkg, "NilClass", obj)

	intClass := r.NewClass(pkg, "Integer", obj)

	r.I64Class = r.NewClass(pkg, "I64", intClass)

	r.NewClass(pkg, "BigInt", intClass)

	r.String = r.NewClass(pkg, "String", obj)

	r.Mirror = r.NewClass(pkg, "ObjectMirror", obj)

	cm := r.NewClass(pkg, "ClassMirror", r.Mirror)

	pm := r.NewClass(pkg, "PackageMirror", r.Mirror)

	r.Package = r.NewClass(pkg, "Package", obj)

	r.Lambda = r.NewClass(pkg, "Lambda", obj)

	r.List = r.NewClass(pkg, "List", obj)

	r.IO = r.NewClass(pkg, "IO", obj)

	initClass(pkg, r.Class)
	initList(pkg, r.List)
	initIO(pkg, r.IO)
	initString(pkg, r.String)
	initI64(pkg, r.I64Class)

	initObjectMirror(r.Mirror)
	initPackageMirror(pkg, pm)
	initClassMirror(pkg, cm)

	for _, f := range inits {
		f(pkg, r)
	}

	initFinalObjectMirror(r.Mirror)

	return nil
}

var inits []func(pkg *Package, r *Registry)

func AddInit(f func(pkg *Package, r *Registry)) {
	inits = append(inits, f)
}
