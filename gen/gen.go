package gen

import (
	"fmt"
	"strings"

	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/insn"
	"github.com/evanphx/m13/value"
)

type Generator struct {
	seq []insn.Instruction

	sp     int
	maxReg int

	locals   map[string]int
	literals []string

	subSequences []*Generator
}

func NewGenerator() (*Generator, error) {
	g := &Generator{
		locals: make(map[string]int),
	}

	return g, nil
}

func (g *Generator) nextReg() int {
	g.sp++

	if g.sp > g.maxReg {
		g.maxReg = g.sp
	}

	return g.sp
}

func (g *Generator) findLiteral(l string) int {
	for i, x := range g.literals {
		if x == l {
			return i
		}
	}

	i := len(g.literals)

	g.literals = append(g.literals, l)

	return i
}

func (g *Generator) Reserve(slot int) {
	g.sp = slot
}

func (g *Generator) Sequence() []insn.Instruction {
	return g.seq
}

func (g *Generator) GenerateTop(gn ast.Node) (*value.Code, error) {
	DesugarAST(gn)

	scope := NewScope()

	err := g.walkScope(gn, scope)
	if err != nil {
		return nil, err
	}

	sc := scope.Close()

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

	code := &value.Code{
		NumRegs:      g.maxReg + 1,
		Instructions: g.seq,
		Literals:     g.literals,
		SubCode:      subs,
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

func DesugarAST(gn ast.Node) {
	ast.Rewrite(gn, func(gn ast.Node) ast.Node {
		switch n := gn.(type) {
		case *ast.Import:
			return &ast.Assign{
				Name: n.Path[len(n.Path)-1],
				Value: &ast.Call{
					Receiver:   &ast.ScopeVar{Name: "LOADER"},
					MethodName: "import",
					Args: []ast.Node{
						&ast.String{Value: strings.Join(n.Path, ".")},
					},
				},
			}
		case *ast.Attribute:
			return &ast.Call{
				Receiver:   n.Receiver,
				MethodName: n.Name,
			}
		default:
			return n
		}
	})
}

func (g *Generator) GenerateScoped(gn ast.Node, scope *ast.Scope) error {
	switch n := gn.(type) {
	case *ast.Import:
		idx := g.findLiteral(strings.Join(n.Path, "."))

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

		idx := g.findLiteral(n.Name)

		g.seq = append(g.seq, insn.Builder.CallOp(g.sp, g.sp, idx))
	case *ast.Call:
		err := g.GenerateScoped(n.Receiver, scope)
		if err != nil {
			return err
		}

		ret := g.sp

		for _, arg := range n.Args {
			g.nextReg()

			err = g.GenerateScoped(arg, scope)
			if err != nil {
				return err
			}
		}

		g.sp = ret

		idx := g.findLiteral(n.MethodName)

		g.seq = append(g.seq, insn.Builder.CallN(g.sp, g.sp, len(n.Args), idx))
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

		idx := g.findLiteral(n.MethodName)

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

		g.seq[patchPos] = insn.Builder.GotoIfFalse(patchSp, len(g.seq))
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
		v, ok := n.Receiver.(*ast.Variable)
		if !ok {
			return fmt.Errorf("Unable to inc type: %T", n.Receiver)
		}

		reg := g.locals[v.Name]

		lit := g.findLiteral("++")

		g.seq = append(g.seq, insn.Builder.Call0(reg, reg, lit))

	case *ast.Dec:
		v, ok := n.Receiver.(*ast.Variable)
		if !ok {
			return fmt.Errorf("Unable to inc type: %T", n.Receiver)
		}

		reg := g.locals[v.Name]

		lit := g.findLiteral("--")

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
		if n.Ref {
			g.seq = append(g.seq, insn.Builder.ReadRef(g.sp, n.Index))
		} else {
			g.seq = append(g.seq, insn.Builder.StoreReg(g.sp, n.Index))
		}

		target := g.sp

		for _, arg := range n.Args {
			g.nextReg()

			err := g.GenerateScoped(arg, scope)
			if err != nil {
				return err
			}
		}

		g.seq = append(g.seq, insn.Builder.Invoke(target, target, len(n.Args)))
		g.sp = target
	case *ast.Lambda:
		sub, err := NewGenerator()
		if err != nil {
			return err
		}

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
		idx := g.findLiteral(n.Name)

		g.seq = append(g.seq, insn.Builder.GetScoped(g.sp, idx))
	case *ast.String:
		idx := g.findLiteral(n.Value)

		g.seq = append(g.seq, insn.Builder.String(g.sp, idx))

	default:
		return fmt.Errorf("Unhandled ast type: %T", gn)
	}

	return nil
}
