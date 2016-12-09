package parser

import (
	"fmt"
	"strings"

	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/lex"
	"github.com/pkg/errors"
)

type Parser struct {
	source string

	root Rule
	expr Rule

	applyDepth int
}

func NewParser(lex *lex.Lexer) (*Parser, error) {
	p := &Parser{source: lex.Source}

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

func (p *Parser) Parse() (ast.Node, error) {
	ml := &markingLexer{r: strings.NewReader(p.source)}

	v, ok := p.root.Match(ml)
	if !ok {
		lines := strings.Split(p.source, "\n")

		var start int64
		var target string
		var targetPos int64

		for _, line := range lines {
			if ml.furthest > start && ml.furthest <= start+int64(len(line)) {
				targetPos = ml.furthest - start - 1
				target = line
				break
			}

			start += (int64(len(line)) + 1)
		}

		marked := fmt.Sprintf("%[1]s\n% [2]*[3]s^", target, targetPos, " ")
		fmt.Printf("%s\n", marked)
		return nil, errors.Wrapf(ErrParse, "Error at position: %d (%d)\n%s", ml.furthest, targetPos, marked)
	}

	return v.(ast.Node), nil
}

func (p *Parser) ParseExpr() (ast.Node, error) {
	ml := &markingLexer{r: strings.NewReader(p.source)}

	v, ok := p.expr.Match(ml)
	if !ok {
		return nil, ErrParse
	}

	return v.(ast.Node), nil
}
