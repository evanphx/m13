package eval

import (
	"github.com/evanphx/m13/gen"
	"github.com/evanphx/m13/lex"
	"github.com/evanphx/m13/parser"
	"github.com/evanphx/m13/value"
	"github.com/evanphx/m13/vm"
)

type Evaluator struct {
}

func NewEvaluator() (*Evaluator, error) {
	return &Evaluator{}, nil
}

func (e *Evaluator) Eval(code string) (value.Value, error) {
	lex, err := lex.NewLexer(code)
	if err != nil {
		return nil, err
	}

	parser, err := parser.NewParser(lex)
	if err != nil {
		return nil, err
	}

	tree, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	g, err := gen.NewGenerator()
	if err != nil {
		return nil, err
	}

	co, err := g.GenerateTop(tree)
	if err != nil {
		return nil, err
	}

	ctx := vm.ExecuteContext{
		Code: co,
	}

	vm, err := vm.NewVM()
	if err != nil {
		return nil, err
	}

	return vm.ExecuteContext(ctx)
}
