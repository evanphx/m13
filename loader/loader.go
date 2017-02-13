package loader

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/gen"
	"github.com/evanphx/m13/parser"
	"github.com/evanphx/m13/value"
)

type Method struct {
	Name string
	Def  *ast.Definition
}

type Package struct {
	name    string
	path    string
	files   []string
	trees   []ast.Node
	methods []*Method
}

func Load(path string) (*Package, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(path)

	lp := &Package{
		name: name,
		path: path,
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(path, file.Name())
		lp.files = append(lp.files, path)

		tree, err := parser.ParseFile(path)
		if err != nil {
			return nil, err
		}

		lp.trees = append(lp.trees, tree)

		lp.scanForMethods(tree)
	}

	return lp, nil
}

func LoadFile(path string) (*Package, error) {
	lp := &Package{
		name: "main",
		path: path,
	}

	lp.files = append(lp.files, path)

	tree, err := parser.ParseFile(path)
	if err != nil {
		return nil, err
	}

	lp.trees = append(lp.trees, tree)

	lp.scanForMethods(tree)

	return lp, nil
}

func (lp *Package) scanForMethods(tree ast.Node) {
	if d, ok := tree.(*ast.Definition); ok {
		lp.methods = append(lp.methods, &Method{
			Name: d.Name,
			Def:  d,
		})

		return
	}

	if blk, ok := tree.(*ast.Block); ok {
		for _, node := range blk.Expressions {
			if d, ok := node.(*ast.Definition); ok {
				lp.methods = append(lp.methods, &Method{
					Name: d.Name,
					Def:  d,
				})
			}
		}
	}
}

func (lp *Package) Methods() []*Method {
	return lp.methods
}

func (lp *Package) Exec(ctx context.Context, env value.Env, r *value.Registry) (*value.Package, error) {
	pkg := r.OpenPackage(lp.name)

	cls := env.MustFindClass("builtin.Loader")

	lo := &Loader{}
	lo.SetClass(cls)
	lo.Search = []string{"lib"}

	ctx = value.SetScoped(ctx, "LOADER", lo)
	ctx = value.SetScoped(ctx, "stdout", value.NewIO(env, os.Stdout))

	for _, tree := range lp.trees {
		g, err := gen.NewGenerator()
		if err != nil {
			return nil, err
		}

		code, err := g.GenerateTop(tree)
		if err != nil {
			return nil, err
		}

		refs := make([]*value.Ref, code.NumRefs)

		for i := 0; i < code.NumRefs; i++ {
			refs[i] = &value.Ref{}
		}

		_, err = env.ExecuteContext(ctx, value.ExecuteContext{
			Code: code,
			Self: pkg,
			Refs: refs,
		})
		if err != nil {
			return nil, err
		}
	}

	/*
		for _, method := range lp.methods {
			g, err := gen.NewGenerator()
			if err != nil {
				return nil, err
			}

			code, err := g.GenerateTop(method.Def)
			if err != nil {
				return nil, err
			}

			pkg.Class(env).AddMethod(&value.MethodDescriptor{
				Name: method.Name,
				Func: func(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
					return env.ExecuteContext(value.ExecuteContext{
						Code: code,
						Self: recv,
						Args: args,
					})
				},
			})
		}
	*/

	return pkg, nil
}

type Loader struct {
	value.Object

	Search []string
}

func Init(pkg *value.Package, r *value.Registry) {
	cls := r.NewClass(pkg, "Loader", r.Object)
	cls.AddMethod(&value.MethodDescriptor{
		Name: "import",
		Signature: value.Signature{
			Required: 1,
		},
		Func: func(ctx context.Context, env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
			lo := recv.(*Loader)
			str := args[0].(*value.String)

			for _, dir := range lo.Search {
				path := filepath.Join(dir, str.String)

				stat, err := os.Stat(path)
				if err != nil {
					continue
				}

				if !stat.IsDir() {
					continue
				}

				lp, err := Load(path)
				if err != nil {
					return nil, err
				}

				return lp.Exec(ctx, env, r)
			}

			return nil, fmt.Errorf("Unable to find %s to import", str.String)
		},
	})
}

func init() {
	value.AddInit(Init)
}
