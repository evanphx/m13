package main

import (
	"context"
	"fmt"
	"os"

	"github.com/evanphx/m13/gen"
	"github.com/evanphx/m13/loader"
	"github.com/evanphx/m13/parser"
	"github.com/evanphx/m13/vm"
)

func oldmain() {
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

func main() {
	lp, err := loader.LoadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	v, err := vm.NewVM()
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()

	_, err = lp.Exec(ctx, v, v.Registry())
	if err != nil {
		panic(err)
	}
}
