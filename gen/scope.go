package gen

import (
	"fmt"

	"github.com/evanphx/m13/ast"
)

type Variable struct {
	Name     string
	NeedsRef bool
	Reads    []*ast.Variable
	Writes   []*ast.Assign
}

type Scope struct {
	Args      []string
	Parent    *Scope
	Variables map[string]*Variable
	Ordered   []*Variable
	Refs      []string
}

func (s *Scope) Find(name string) *Variable {
	if v, ok := s.Variables[name]; ok {
		return v
	}

	if s.Parent != nil {
		return s.Parent.Find(name)
	}

	return nil
}

func (s *Scope) addRef(name string) int {
	for i := 0; i < len(s.Refs); i++ {
		if s.Refs[i] == name {
			return i
		}
	}

	i := len(s.Refs)

	s.Refs = append(s.Refs, name)

	return i
}

func (s *Scope) findRef(name string) int {
	for i := 0; i < len(s.Refs); i++ {
		if s.Refs[i] == name {
			return i
		}
	}

	panic(fmt.Sprintf("unknown ref: '%s'", name))
}

func (s *Scope) findArg(name string) int {
	for i := 0; i < len(s.Args); i++ {
		if s.Args[i] == name {
			return i
		}
	}

	return -1
}

func (s *Scope) makeRef(name string) {
	s.addRef(name)
	if s.Parent != nil {
		s.Parent.makeRef(name)
	}
}

func (s *Scope) SetArgs(args []string) {
	s.Args = args

	for _, name := range args {
		v := &Variable{
			Name: name,
		}

		s.Variables[name] = v

		s.Ordered = append(s.Ordered, v)
	}
}

func (s *Scope) Read(n *ast.Variable) {
	name := n.Name

	var (
		v  *Variable
		ok bool
	)

	if v, ok = s.Variables[name]; ok {
		v.Reads = append(v.Reads, n)
	} else {
		if s.Parent != nil {
			if pv := s.Parent.Find(name); pv != nil {
				v = &Variable{
					Name:  name,
					Reads: []*ast.Variable{n},
				}

				s.Variables[name] = v

				s.Ordered = append(s.Ordered, v)

				pv.NeedsRef = true
				v.NeedsRef = true
				s.makeRef(name)
			}
		} else {
			panic(fmt.Sprintf("reading unassigned variable: %s", name))
		}
	}

	/*
		else {
			v = &Variable{
				Name:  name,
				Reads: []*ast.Variable{n},
			}

			s.Variables[name] = v

			s.Ordered = append(s.Ordered, v)
		}
	*/

}

func (s *Scope) Write(n *ast.Assign) {
	name := n.Name

	if v, ok := s.Variables[name]; ok {
		v.Writes = append(v.Writes, n)
	} else {
		v := &Variable{
			Name:   name,
			Writes: []*ast.Assign{n},
		}

		s.Variables[name] = v

		s.Ordered = append(s.Ordered, v)
	}

	if s.Parent != nil {
		if v := s.Parent.Find(name); v != nil {
			v.NeedsRef = true
			s.makeRef(name)
		}
	}
}

func (s *Scope) Close() *ast.Scope {
	locals := len(s.Args)

	sc := &ast.Scope{
		Refs: s.Refs,
	}

	for _, v := range s.Ordered {
		if v.NeedsRef {
			ref := s.findRef(v.Name)

			for _, u := range v.Reads {
				u.Ref = true
				u.Index = ref
			}

			for _, u := range v.Writes {
				u.Ref = true
				u.Index = ref
			}
		} else {
			idx := s.findArg(v.Name)
			if idx == -1 {
				idx = locals
				locals++
			}

			sc.Locals = append(sc.Locals, v.Name)

			for _, u := range v.Reads {
				u.Index = idx
			}

			for _, u := range v.Writes {
				u.Index = idx
			}
		}
	}

	return sc
}

func NewScope() *Scope {
	return &Scope{
		Variables: make(map[string]*Variable),
	}
}

func (g *Generator) walkScopeOld(gn ast.Node, scope *Scope) error {
	ast.Descend(gn, func(dn ast.Node) bool {
		switch n := dn.(type) {
		case *ast.Assign:
			scope.Write(n)
		case *ast.Variable:
			scope.Read(n)
		case *ast.Lambda:
			subScope := NewScope()
			subScope.Parent = scope

			subScope.SetArgs(n.Args)

			g.walkScope(n.Expr, subScope)

			n.Scope = subScope.Close()

			return false
		}

		return true
	})

	return nil
}

func (g *Generator) walkScope(gn ast.Node, scope *Scope) error {
	type scopeWork struct {
		lam   *ast.Lambda
		scope *Scope
	}

	var work, done []*scopeWork

	lam := &ast.Lambda{Expr: gn}

	work = append(work, &scopeWork{lam, scope})

	for len(work) > 0 {
		ls := work[0]
		work = work[1:]

		ast.Descend(ls.lam.Expr, func(dn ast.Node) bool {
			switch n := dn.(type) {
			case *ast.Variable:
				ls.scope.Read(n)
			case *ast.Assign:
				ls.scope.Write(n)
			case *ast.Lambda:
				subScope := NewScope()
				subScope.Parent = ls.scope
				subScope.SetArgs(n.Args)

				work = append(work, &scopeWork{n, subScope})

				return false
			}

			return true
		})

		done = append(done, ls)
	}

	for _, ls := range done {
		ls.lam.Scope = ls.scope.Close()
	}

	return nil
}
