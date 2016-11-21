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

	panic(fmt.Sprintf("unknown ref: %s", name))
}

func (s *Scope) makeRef(name string) {
	s.addRef(name)
	if s.Parent != nil {
		s.Parent.makeRef(name)
	}
}

func (s *Scope) Read(n *ast.Variable) {
	name := n.Name

	if v, ok := s.Variables[name]; ok {
		v.Reads = append(v.Reads, n)
	} else {
		v := &Variable{
			Reads: []*ast.Variable{n},
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

func (s *Scope) Write(n *ast.Assign) {
	name := n.Name

	if v, ok := s.Variables[name]; ok {
		v.Writes = append(v.Writes, n)
	} else {
		v := &Variable{
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
	locals := 0

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
			sc.Locals = append(sc.Locals, v.Name)

			for _, u := range v.Reads {
				u.Index = locals
			}

			for _, u := range v.Writes {
				u.Index = locals
			}

			locals++
		}
	}

	return sc
}

func NewScope() *Scope {
	return &Scope{
		Variables: make(map[string]*Variable),
	}
}

func (g *Generator) walkScope(gn ast.Node, scope *Scope) error {
	switch n := gn.(type) {
	case *ast.Op:
		err := g.walkScope(n.Left, scope)
		if err != nil {
			return err
		}

		err = g.walkScope(n.Right, scope)
		if err != nil {
			return err
		}
	case *ast.Block:
		for _, ex := range n.Expressions {
			err := g.walkScope(ex, scope)
			if err != nil {
				return err
			}
		}
	case *ast.If:
		err := g.walkScope(n.Cond, scope)
		if err != nil {
			return err
		}

		err = g.walkScope(n.Body, scope)
		if err != nil {
			return err
		}
	case *ast.While:
		err := g.walkScope(n.Cond, scope)
		if err != nil {
			return err
		}

		err = g.walkScope(n.Body, scope)
		if err != nil {
			return err
		}
	case *ast.Inc:
		err := g.walkScope(n.Receiver, scope)
		if err != nil {
			return err
		}
	case *ast.Dec:
		err := g.walkScope(n.Receiver, scope)
		if err != nil {
			return err
		}
	case *ast.Assign:
		err := g.walkScope(n.Value, scope)
		if err != nil {
			return err
		}

		scope.Write(n)
	case *ast.Variable:
		scope.Read(n)
	case *ast.Lambda:
		subScope := NewScope()
		subScope.Parent = scope

		err := g.walkScope(n.Expr, subScope)
		if err != nil {
			return err
		}

		n.Scope = subScope.Close()
	}

	return nil
}
