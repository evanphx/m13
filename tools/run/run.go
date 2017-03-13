package main

import (
	"context"
	"os"

	"github.com/evanphx/m13/loader"
	"github.com/evanphx/m13/vm"
)

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
