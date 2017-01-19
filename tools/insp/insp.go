package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"

	"github.com/evanphx/m13/gen"
	"github.com/evanphx/m13/parser"
)

var (
	fParse   = flag.String("parse", "", "file to parse and show output of")
	fDesugar = flag.Bool("desugar", false, "run the ast through desugaring")
	fCompile = flag.String("compile", "", "file to compile and show output of")
)

func main() {
	flag.Parse()

	if *fParse != "" {
		node, err := parser.ParseFile(*fParse)
		if err != nil {
			fmt.Printf("Error parsing '%s': %s\n", *fParse, err)
			return
		}

		if *fDesugar {
			node = gen.DesugarAST(node)
		}

		ast.Print(nil, node)

		return
	}

	if *fCompile != "" {
		tree, err := parser.ParseFile(*fCompile)
		if err != nil {
			fmt.Printf("Error parsing '%s': %s\n", *fCompile, err)
			return
		}

		g, err := gen.NewGenerator()
		if err != nil {
			fmt.Printf("Error creating generator\n")
			return
		}

		co, err := g.GenerateTop(tree)
		if err != nil {
			fmt.Printf("Error generating '%s': %s\n", *fCompile, err)
			return
		}

		co.Disassemble(os.Stdout)
		return
	}

	flag.Usage()
}
