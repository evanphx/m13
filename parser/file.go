package parser

import (
	"io/ioutil"

	"github.com/evanphx/m13/ast"
)

func ParseFile(path string) (ast.Node, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p, err := NewParser(string(data))
	if err != nil {
		return nil, err
	}

	return p.Parse()
}
