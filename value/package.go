package value

// m13
type Package struct {
	Object

	Name    string
	Classes map[string]*Class
}

// m13
type PackageMirror struct {
	Object

	p *Package
}

func (m *PackageMirror) AddMethod(env Env, name *String, l *Lambda) (Value, error) {
	m.p.Class(env).AddMethod(&MethodDescriptor{
		Name: name.String,
		Func: func(env Env, recv Value, args []Value) (Value, error) {
			return env.ExecuteContext(ExecuteContext{
				Code: l.Code,
				Refs: l.Refs,
				Args: args,
			})
		},
	})

	return nil, nil
}
