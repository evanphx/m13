package value

func TrueClass(env Env, v Value) *Class {
	cls := v.Class(env)

	for cls.metaclass {
		cls = cls.Parent
	}

	return cls
}

func (c *Class) AddMethod(cfg *MethodDescriptor) {
	c.Methods[cfg.Name] = &Method{
		Name:      cfg.Name,
		Signature: cfg.Signature,
		F:         cfg.Func,
	}
}
