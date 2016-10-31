package parser

import (
	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/lex"
	"github.com/pkg/errors"
)

type Parser struct {
	lex *lex.Lexer

	root Rule
}

func NewParser(lex *lex.Lexer) (*Parser, error) {
	p := &Parser{lex: lex}

	p.SetupRules()

	return p, nil
}

var ErrParse = errors.New("parse error")

type NodeStack struct {
	stack []ast.Node
}

func (s *NodeStack) Push(v ast.Node) {
	s.stack = append(s.stack, v)
}

func (s *NodeStack) Pop() ast.Node {
	v := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return v
}

func (p *Parser) expect(t lex.Type) (*lex.Value, error) {
	ne, err := p.lex.Next()
	if err != nil {
		return nil, err
	}

	if ne.Type != t {
		return nil, errors.Wrapf(ErrParse,
			"expected %d, got %d", t, ne.Type)
	}

	return ne, nil
}

func (p *Parser) Parse() (ast.Node, error) {
	ml := &markingLexer{lex: p.lex}

	v, ok := p.root.Match(ml)
	if !ok {
		return nil, ErrParse
	}

	return v.(ast.Node), nil
}

/*
func (p *Parser) oldParse() (ast.Node, error) {
	var stack NodeStack

	for {
		le, err := p.lex.Next()
		if err != nil {
			return nil, err
		}

		switch le.Type {
		case lex.Integer:
			stack.Push(&ast.Integer{le.Value.(int64)})
		case lex.String:
			stack.Push(&ast.String{le.Value.(string)})
		case lex.Atom:
			stack.Push(&ast.Atom{le.Value.(string)})
		case lex.True:
			stack.Push(&ast.True{})
		case lex.False:
			stack.Push(&ast.False{})
		case lex.Nil:
			stack.Push(&ast.Nil{})
		case lex.Word:
			stack.Push(&ast.Variable{Name: le.Value.(string)})
		case lex.Dot:
			ne, err := p.expect(lex.Word)
			if err != nil {
				return nil, err
			}

			call := &ast.Call{
				Receiver:   stack.Pop(),
				MethodName: ne.Value.(string),
			}

			_, err = p.expect(lex.OpenParen)
			if err != nil {
				return nil, err
			}

			stack.Push(&ast.Call{
				Receiver:   stack.Pop(),
				MethodName: ne.Value.(string),
			})
		case lex.Term:
			return stack.Pop(), nil
		default:
			return nil, errors.Wrapf(ErrParse, "unexpected lexem")
		}
	}

	return nil, nil
}
*/
