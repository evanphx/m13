package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/imports"
)

var (
	fType = flag.String("T", "", "type to generate")
)

func main() {
	dir := os.Args[1]
	var conf loader.Config

	conf.TypeCheckFuncBodies = func(_ string) bool { return false }
	conf.TypeChecker.DisableUnusedImportCheck = true
	conf.TypeChecker.Importer = importer.Default()
	conf.ParserMode = parser.ParseComments

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var (
		astfiles []*ast.File
	)

	for _, fi := range files {
		if filepath.Ext(fi.Name()) != ".go" {
			continue
		}

		path := filepath.Join(dir, fi.Name())
		f, err := conf.ParseFile(path, nil)
		if err != nil {
			log.Fatal(err)
		}

		astfiles = append(astfiles, f)
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	conf.CreateFromFiles(abs, astfiles...)

	_, err = conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer

	gen(astfiles, &buf)

	opt := &imports.Options{Comments: true}
	theBytes := buf.Bytes()

	res, err := imports.Process("m13.gen.go", theBytes, opt)
	if err != nil {
		log.Print(err)
		res = theBytes
	}

	err = ioutil.WriteFile(filepath.Join(dir, "m13.gen.go"), res, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

type arg struct {
	Name   string
	GoType string
}

type method struct {
	Name     string
	GoName   string
	GenName  string
	NumArgs  int
	RecvType string
	Args     []*arg
	Aliases  []string

	IncludeEnv bool
}

func (m *method) Parse(text string) {
	parts := strings.Split(strings.TrimSpace(text), " ")

	for _, part := range parts {
		eq := strings.IndexByte(part, '=')
		if eq != -1 {
			val := part[eq+1:]

			switch part[:eq] {
			case "name":
				m.Name = val
			case "alias":
				m.Aliases = strings.Split(val, ",")
			default:
				log.Fatalf("Unknown key: %s", part[:eq])
			}
		}
	}
}

type exportedType struct {
	GlobalName string
	GoName     string
	Name       string
	Parent     string

	Methods []*method
}

func (et *exportedType) Parse(text string) {
	parts := strings.Split(strings.TrimSpace(text), " ")

	for _, part := range parts {
		eq := strings.IndexByte(part, '=')
		if eq != -1 {
			val := part[eq+1:]

			switch part[:eq] {
			case "name":
				et.Name = val
			case "parent":
				et.Parent = val
			default:
				log.Fatalf("Unknown key: %s", part[:eq])
			}
		}
	}
}

type glue struct {
	Package string
	Types   []*exportedType
}

const codeTemplate2 = `

{{ $pkg := .Package }}

{{range $type := .Types}}
	{{range .Methods}}
		func {{.GenName}}_adapter(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
			if len(args) != {{.NumArgs}} {
				return env.ArgumentError(len(args), {{.NumArgs}})
			}

			self := recv.({{.RecvType}})

			{{range $i, $e := .Args}}
				a{{$i}} := args[{{$i}}].({{$e.GoType}})
			{{end}}

			ret, err := self.{{.GoName}}(
				{{if .IncludeEnv}}
				env,
				{{end}}
				{{range $i, $e := .Args}}
					a{{$i}},
				{{end}}
			)

			if err != nil {
				return nil, err
			}

			return ret, nil
		}

		var method_desc_{{.GenName}} = &value.MethodDescriptor {
			Name: "{{.Name}}",
			Aliases: []string{
				{{range .Aliases}}
					"{{.}}",
				{{end}}
			},
			Signature: value.Signature{
				Required: {{.NumArgs}},
				Args: []string {
					{{range .Args}}
						"{{.Name}}",
					{{end}}
				},
			},
			Func: {{.GenName}}_adapter,
		}
	{{end}}

	var methods_{{.Name}} = []*value.MethodDescriptor{
			{{range .Methods}}
				method_desc_{{.GenName}},
			{{end}}
		}

	func setup_{{.Name}}(setup value.Setup) {
		setup.ApplyMethods("{{$pkg}}.{{.Name}}", methods_{{.Name}})
	/*
		pkg := setup.OpenPackage("{{$pkg}}")

		methods := make(map[string]*value.Method)

		{{range .Methods}}
			methods["{{.Name}}"] = setup.MakeMethod(&value.MethodConfig{
				Name: "{{.Name}}",
				Func: {{.GenName}}_adapter,
			})
		{{end}}

		setup.MakeClass(&value.ClassConfig{
			Package: pkg,
			Name: "{{.Name}}",
			Parent: "{{.Parent}}",
			Methods: methods,
			GlobalName: "{{.GlobalName}}",
		})
	*/
	}

	var _ = value.RegisterSetup(setup_{{.Name}})
{{end}}
`

func gen(files []*ast.File, out io.Writer) {
	out.Write([]byte("package builtin\n"))

	for _, f := range files {
		genFile(f, out)
	}
}

func genFile(f *ast.File, out io.Writer) {
	var export []*exportedType

	byName := map[string]*exportedType{}

	for _, decl := range f.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok {
			if gd.Doc.Text() == "" {
				continue
			}

			for _, spec := range gd.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					name := ts.Name.String()

					et := &exportedType{
						GlobalName: fmt.Sprintf("builtin.%s", name),
						GoName:     name,
						Name:       name,
					}

					if strings.HasPrefix(gd.Doc.Text(), "m13 ") {
						et.Parse(gd.Doc.Text()[4:])
					}

					byName[name] = et
					export = append(export, et)
				}
			}
		}
	}

	for _, decl := range f.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			if fd.Recv == nil {
				continue
			}

			if !strings.HasPrefix(fd.Doc.Text(), "m13") {
				continue
			}

			typ := fd.Recv.List[0].Type

			var (
				recv     string
				recvType string
			)

			if ident, ok := typ.(*ast.Ident); ok {
				recv = ident.Name
				recvType = ident.Name
			} else if star, ok := typ.(*ast.StarExpr); ok {
				recv = star.X.(*ast.Ident).Name
				recvType = "*" + recv
			}

			name := fd.Name.Name

			var (
				args       []*arg
				includeEnv bool
			)

			for _, field := range fd.Type.Params.List {
				if field.Names[0].Name == "env" {
					includeEnv = true
					continue
				}

				for _, name := range field.Names {
					if id, ok := field.Type.(*ast.Ident); ok {
						args = append(args, &arg{Name: name.Name, GoType: id.Name})
					} else if st, ok := field.Type.(*ast.StarExpr); ok {
						args = append(args, &arg{Name: name.Name, GoType: "*" + st.X.(*ast.Ident).Name})
					}
				}
			}

			t, ok := byName[recv]
			if !ok {
				panic(fmt.Sprintf("where is %s", recv))
			}

			meth := &method{
				GoName:     name,
				Name:       name,
				GenName:    fmt.Sprintf("%s_%s", recv, name),
				NumArgs:    len(args),
				RecvType:   recvType,
				Args:       args,
				IncludeEnv: includeEnv,
			}

			if strings.HasPrefix(fd.Doc.Text(), "m13 ") {
				meth.Parse(fd.Doc.Text()[4:])
			}

			t.Methods = append(t.Methods, meth)
		}
	}

	var g glue
	g.Package = "builtin"
	g.Types = export

	t, err := template.New("code").Parse(codeTemplate2)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(out, &g)
	if err != nil {
		log.Fatal(err)
	}
}
