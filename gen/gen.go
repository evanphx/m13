package gen

import (
	"fmt"
	"strings"

	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/insn"
	"github.com/evanphx/m13/value"
)

type Generator struct {
	env  value.Env
	name string
	seq  []insn.Instruction

	sp     int
	maxReg int

	scope *ast.Scope

	calls   []*value.CallSite
	strings []*value.String

	subSequences []*Generator

	signature *value.Signature
}

func NewGenerator(env value.Env, name string) (*Generator, error) {
	g := &Generator{env: env, name: name}

	return g, nil
}

func (g *Generator) nextReg() int {
	g.sp++

	if g.sp > g.maxReg {
		g.maxReg = g.sp
	}

	return g.sp
}

func (g *Generator) a(i insn.Instruction) {
	g.seq = append(g.seq, i)
}

func (g *Generator) findString(l string) int {
	for i, x := range g.strings {
		if x.String == l {
			return i
		}
	}

	i := len(g.strings)

	g.strings = append(g.strings, g.env.InternString(l))

	return i
}

func (g *Generator) addCallsite(l string) (*value.CallSite, int) {
	cs := &value.CallSite{
		Name: l,
	}

	i := len(g.calls)

	g.calls = append(g.calls, cs)

	return cs, i
}

func (g *Generator) Reserve(slot int) {
	g.sp = slot
}

func (g *Generator) Sequence() []insn.Instruction {
	return g.seq
}

func (g *Generator) GenerateTop(gn ast.Node) (*value.Code, error) {
	gn = DesugarAST(gn)

	scope := NewScope()

	err := g.walkScope(gn, scope)
	if err != nil {
		return nil, err
	}

	sc := scope.Close()

	g.scope = sc

	g.sp += len(sc.Locals)
	g.maxReg = g.sp

	err = g.GenerateScoped(gn, sc)
	if err != nil {
		return nil, err
	}

	g.seq = append(g.seq, insn.Builder.Return(len(sc.Locals)))

	return g.Code()
}

func (g *Generator) Code() (*value.Code, error) {
	var subs []*value.Code

	for _, sg := range g.subSequences {
		c, err := sg.Code()
		if err != nil {
			return nil, err
		}

		subs = append(subs, c)
	}

	var p PeepholeOptz

	p.Optimize(g.seq)

	code := &value.Code{
		Name:         g.name,
		NumRegs:      g.maxReg + 1,
		NumRefs:      len(g.scope.Refs),
		Instructions: g.seq,
		Strings:      g.strings,
		Calls:        g.calls,
		SubCode:      subs,
		Signature:    g.signature,
	}

	return code, nil
}

func (g *Generator) Generate(gn ast.Node) error {
	scope := NewScope()

	err := g.walkScope(gn, scope)
	if err != nil {
		return err
	}

	sc := scope.Close()

	g.sp += len(sc.Locals)
	g.maxReg = g.sp

	err = g.GenerateScoped(gn, sc)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) GenerateLambda(gn ast.Node, sc *ast.Scope) error {
	g.sp += len(sc.Locals)
	g.maxReg = g.sp

	err := g.GenerateScoped(gn, sc)
	if err != nil {
		return err
	}

	g.seq = append(g.seq, insn.Builder.Return(len(sc.Locals)))

	return nil
}

func DesugarAST(gn ast.Node) ast.Node {
	return ast.Rewrite(gn, func(gn ast.Node) ast.Node {
		switch n := gn.(type) {
		case *ast.Import:
			if n.Relative {
				return &ast.Assign{
					Name: n.Path[len(n.Path)-1],
					Value: &ast.Call{
						Receiver:   &ast.ScopeVar{Name: "LOADER"},
						MethodName: "import_relative",
						Args: &ast.Args{
							Args: []ast.Node{
								&ast.String{Value: strings.Join(n.Path, ".")},
							},
						},
					},
				}
			} else {
				return &ast.Assign{
					Name: n.Path[len(n.Path)-1],
					Value: &ast.Call{
						Receiver:   &ast.ScopeVar{Name: "LOADER"},
						MethodName: "import",
						Args: &ast.Args{
							Args: []ast.Node{
								&ast.String{Value: strings.Join(n.Path, ".")},
							},
						},
					},
				}
			}
		case *ast.Definition:
			elements := []ast.Node{}

			elements = append(elements,
				&ast.UpCall{
					Receiver:   &ast.Self{},
					MethodName: "add_method",
					Args: []ast.Node{
						&ast.String{Value: n.Name.Name},
						&ast.Lambda{
							Name: n.Name.Name,
							Args: n.Arguments,
							Expr: n.Body,
						},
					},
				},
			)

			if n.Name.Operator != "" {
				elements = append(elements,
					&ast.UpCall{
						Receiver:   &ast.Self{},
						MethodName: "alias_method",
						Args: []ast.Node{
							&ast.String{Value: n.Name.Name},
							&ast.String{Value: n.Name.Operator},
						},
					},
				)
			}

			return &ast.Block{Expressions: elements}
		case *ast.ClassDefinition:
			return &ast.Assign{
				Name: n.Name,
				Value: &ast.UpCall{
					Receiver:   &ast.Self{},
					MethodName: "add_class",
					Args: []ast.Node{
						&ast.String{Value: n.Name},
						&ast.Lambda{
							Name: n.Name + ".__body__",
							Expr: n.Body,
						},
					},
				},
			}
		case *ast.Has:
			var traits []ast.Node

			for _, t := range n.Traits {
				traits = append(traits, &ast.String{Value: t})
			}

			return &ast.UpCall{
				Receiver:   &ast.Self{},
				MethodName: "add_ivar",
				Args: []ast.Node{
					&ast.String{Value: n.Variable},
					&ast.List{Elements: traits},
				},
			}
		case *ast.Attribute:
			return &ast.Call{
				Receiver:   n.Receiver,
				MethodName: n.Name,
				Args:       &ast.Args{},
			}
		default:
			return n
		}
	})
}

func (g *Generator) GenerateScoped(gn ast.Node, scope *ast.Scope) error {
	switch n := gn.(type) {
	case *ast.Import:
		idx := g.findString(strings.Join(n.Path, "."))

		g.seq = append(g.seq, insn.Builder.GetScoped(g.sp, idx))
	case *ast.Self:
		g.seq = append(g.seq, insn.Builder.Self(g.sp))
	case *ast.Integer:
		g.seq = append(g.seq, insn.Builder.Store(g.sp, insn.Int(n.Value)))
	case *ast.Op:
		err := g.GenerateScoped(n.Left, scope)
		if err != nil {
			return err
		}

		g.nextReg()

		err = g.GenerateScoped(n.Right, scope)
		if err != nil {
			return err
		}

		g.sp--

		_, idx := g.addCallsite(n.Name)

		g.seq = append(g.seq, insn.Builder.CallOp(g.sp, g.sp, idx))
	case *ast.Call:
		err := g.GenerateScoped(n.Receiver, scope)
		if err != nil {
			return err
		}

		ret := g.sp

		var (
			pos   int
			named int
		)

		cs, idx := g.addCallsite(n.MethodName)

		for _, arg := range n.Args.Args {
			g.nextReg()

			if na, ok := arg.(*ast.NamedArg); ok {
				named++

				err = g.GenerateScoped(na.Value, scope)
				if err != nil {
					return err
				}

				cs.KWTable = append(cs.KWTable, na.Name)
			} else {
				if named > 0 {
					return fmt.Errorf("Positional args after named are not supported")
				}

				pos++

				err = g.GenerateScoped(arg, scope)
				if err != nil {
					return err
				}
			}
		}

		g.sp = ret

		if named == 0 {
			g.seq = append(
				g.seq,
				insn.Builder.CallN(g.sp, g.sp, pos, idx))
		} else {
			g.seq = append(
				g.seq,
				insn.Builder.CallKW(g.sp, g.sp, pos, named, idx))
		}

	case *ast.UpCall:
		err := g.GenerateScoped(n.Receiver, scope)
		if err != nil {
			return err
		}

		g.seq = append(g.seq, insn.Builder.GetMirror(g.sp, g.sp))

		ret := g.sp

		for _, arg := range n.Args {
			g.nextReg()

			err = g.GenerateScoped(arg, scope)
			if err != nil {
				return err
			}
		}

		g.sp = ret

		_, idx := g.addCallsite(n.MethodName)

		g.seq = append(g.seq, insn.Builder.CallN(g.sp, g.sp, len(n.Args), idx))
	case *ast.Block:
		for _, ex := range n.Expressions {
			err := g.GenerateScoped(ex, scope)
			if err != nil {
				return err
			}
		}
	case *ast.If:
		err := g.GenerateScoped(n.Cond, scope)
		if err != nil {
			return err
		}

		patchSp := g.sp

		patchPos := len(g.seq)

		g.seq = append(g.seq, insn.Builder.GotoIfFalse(patchSp, 0))

		err = g.GenerateScoped(n.Body, scope)
		if err != nil {
			return err
		}

		if n.Else != nil {
			ppPos := len(g.seq)
			g.a(insn.Builder.Goto(0))

			g.seq[patchPos] = insn.Builder.GotoIfFalse(patchSp, len(g.seq))

			err = g.GenerateScoped(n.Else, scope)
			if err != nil {
				return err
			}

			g.seq[ppPos] = insn.Builder.Goto(len(g.seq))
		} else {
			g.seq[patchPos] = insn.Builder.GotoIfFalse(patchSp, len(g.seq))
		}

	case *ast.While:
		condPos := len(g.seq)

		err := g.GenerateScoped(n.Cond, scope)
		if err != nil {
			return err
		}

		patchSp := g.sp

		patchPos := len(g.seq)

		g.seq = append(g.seq, insn.Builder.GotoIfFalse(patchSp, 0))

		err = g.GenerateScoped(n.Body, scope)
		if err != nil {
			return err
		}

		g.seq = append(g.seq, insn.Builder.Goto(condPos))

		g.seq[patchPos] = insn.Builder.GotoIfFalse(patchSp, len(g.seq))

	case *ast.Inc:
		err := g.GenerateScoped(n.Receiver, scope)
		if err != nil {
			return err
		}

		reg := g.sp

		_, lit := g.addCallsite("++")

		g.seq = append(g.seq, insn.Builder.Call0(reg, reg, lit))

	case *ast.Dec:
		err := g.GenerateScoped(n.Receiver, scope)
		if err != nil {
			return err
		}

		reg := g.sp

		_, lit := g.addCallsite("--")

		g.seq = append(g.seq, insn.Builder.Call0(reg, reg, lit))

	case *ast.Assign:
		err := g.GenerateScoped(n.Value, scope)
		if err != nil {
			return err
		}

		if n.Ref {
			g.seq = append(g.seq, insn.Builder.StoreRef(n.Index, g.sp))
		} else {
			g.seq = append(g.seq, insn.Builder.StoreReg(n.Index, g.sp))
		}
	case *ast.Variable:
		if n.Ref {
			g.seq = append(g.seq, insn.Builder.ReadRef(g.sp, n.Index))
		} else {
			g.seq = append(g.seq, insn.Builder.StoreReg(g.sp, n.Index))
		}
	case *ast.Invoke:
		err := g.GenerateScoped(n.Var, scope)
		if err != nil {
			return err
		}

		target := g.sp

		for _, arg := range n.Args.Args {
			g.nextReg()

			err := g.GenerateScoped(arg, scope)
			if err != nil {
				return err
			}
		}

		g.seq = append(g.seq, insn.Builder.Invoke(target, target, len(n.Args.Args)))
		g.sp = target
	case *ast.Lambda:
		sub, err := NewGenerator(g.env, n.Name)
		if err != nil {
			return err
		}

		var sig value.Signature

		sig.Required = len(n.Args)

		for _, arg := range n.Args {
			sig.Args = append(sig.Args, arg.Name)
		}

		sub.signature = &sig

		sub.scope = n.Scope

		err = sub.GenerateLambda(n.Expr, n.Scope)
		if err != nil {
			return err
		}

		pos := len(g.subSequences)
		g.subSequences = append(g.subSequences, sub)

		g.seq = append(g.seq, insn.Builder.CreateLambda(g.sp, len(n.Args), len(n.Scope.Refs), pos))
		for _, name := range n.Scope.Refs {
			parentPos := scope.RefIndex(name)
			g.seq = append(g.seq, insn.Builder.ReadRef(0, parentPos))
		}

	case *ast.ScopeVar:
		idx := g.findString(n.Name)

		g.seq = append(g.seq, insn.Builder.GetScoped(g.sp, idx))
	case *ast.String:
		idx := g.findString(n.Value)

		g.seq = append(g.seq, insn.Builder.String(g.sp, idx))

	case *ast.List:
		listReg := g.sp
		g.a(insn.Builder.NewList(listReg, len(n.Elements)))

		g.nextReg()

		for _, e := range n.Elements {
			g.GenerateScoped(e, scope)
			g.a(insn.Builder.ListAppend(listReg, g.sp))
		}

		g.sp--
	case *ast.Map:
		mapReg := g.sp
		g.a(insn.Builder.NewMap(mapReg))

		for _, pair := range n.Elements {
			g.nextReg()
			keySp := g.sp

			g.GenerateScoped(pair.Key, scope)

			g.nextReg()

			g.GenerateScoped(pair.Value, scope)

			g.a(insn.Builder.SetMap(mapReg, keySp))

			g.sp -= 2
		}
	case *ast.IVar:
		idx := g.findString(n.Name)

		g.a(insn.Builder.GetIvar(g.sp, idx))

	case *ast.IVarAssign:
		idx := g.findString(n.Name)

		err := g.GenerateScoped(n.Value, scope)
		if err != nil {
			return err
		}

		g.a(insn.Builder.SetIvar(g.sp, idx))

	default:
		return fmt.Errorf("Unhandled ast type: %T (%#v)", gn, gn)
	}

	return nil
}
