package main

import (
	"fmt"
	"os"

	"github.com/evanphx/m13/gen"
	"github.com/evanphx/m13/parser"
)

func main() {
	tree, err := parser.ParseFile(os.Args[1])
	if err != nil {
		fmt.Printf("ERROR: Unable to load file: %s\n", err)
		os.Exit(1)
	}

	g, err := gen.NewGenerator()
	if err != nil {
		fmt.Printf("ERROR: Unable to create generator: %s\n", err)
		os.Exit(1)
	}

	code, err := g.GenerateTop(tree)
	if err != nil {
		fmt.Printf("ERROR: Unable to code gen: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("code: %v\n", code)
}
