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

	intClass := r.NewClass(pkg, "Integer", obj)

	r.I64Class = r.NewClass(pkg, "I64", intClass)

	r.NewClass(pkg, "BigInt", intClass)

	r.NewClass(pkg, "String", obj)

	r.Mirror = r.NewClass(pkg, "ObjectMirror", obj)

	return nil
}
