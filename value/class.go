package value

func TrueClass(env Env, v Value) *Class {
	cls := v.Class(env)

	for cls.metaclass {
		cls = cls.Parent
	}

	return cls
}

func (c *Class) AddMethod(cfg *MethodDescriptor) {
	method := &Method{
		Name:      cfg.Name,
		Signature: cfg.Signature,
		F:         cfg.Func,
	}

	c.Methods[cfg.Name] = method

	for _, arg := range cfg.Aliases {
		c.Methods[arg] = method
	}
}
