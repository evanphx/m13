package main

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/loader"
)

func main() {
	path := os.Args[2]
	var conf loader.Config

	conf.TypeCheckFuncBodies = func(_ string) bool { return false }
	conf.TypeChecker.DisableUnusedImportCheck = true
	conf.TypeChecker.Importer = importer.Default()
	conf.ParserMode = parser.ParseComments

	f, err := conf.ParseFile(path, nil)
	if err != nil {
		log.Fatal(err)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	conf.CreateFromFiles(abs, f)

	_, err = conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	gen(f)
}

type arg struct {
	GoType string
}

type method struct {
	Name     string
	NumArgs  int
	RecvType string
	Args     []*arg
}

type exportedType struct {
	GoName string
	Name   string
	Parent string

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

const codeTemplate2 = `package {{.Package}}

{{range $type := .Types}}
	var type_{{.Name}} *value.Type

	func (_ {{.GoName}}) Type() *value.Type {
		return type_{{.Name}}
	}

	{{range .Methods}}
		func {{.Name}}_adapter(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
			if len(args) != {{.NumArgs}} {
				return env.ArgumentError(len(args), {{.NumArgs}})
			}

			self := recv.({{.RecvType}})

			{{range $i, $e := .Args}}
				a{{$i}} := args[{{$i}}].({{$e.GoType}})
			{{end}}

			ret, err := self.{{.Name}}(
				{{range $i, $e := .Args}}
					a{{$i}},
				{{end}}
			)

			if err != nil {
				return nil, err
			}

			return ret, nil
		}
	{{end}}
{{end}}

func init() {
	pkg := value.OpenPackage("{{.Package}}")

	var methods map[string]*value.Method

	{{range .Types}}
		methods = make(map[string]*value.Method)

		{{range .Methods}}
			methods["{{.Name}}"] = value.MakeMethod(&value.MethodConfig{
				Name: "{{.Name}}",
				Func: {{.Name}}_adapter,
			})
		{{end}}

		type_{{.Name}} = value.MakeType(&value.TypeConfig{
			Package: pkg,
			Name: "{{.Name}}",
			Parent: "{{.Parent}}",
			Methods: methods,
		})
	{{end}}
}
`

func gen(f *ast.File) {
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

					et := &exportedType{GoName: name, Name: name}

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

			recv := fd.Recv.List[0].Type.(*ast.Ident).Name
			name := fd.Name.Name

			var args []*arg

			for _, field := range fd.Type.Params.List {
				args = append(args, &arg{GoType: field.Type.(*ast.Ident).Name})
			}

			t := byName[recv]

			t.Methods = append(t.Methods, &method{
				Name:     name,
				NumArgs:  len(args),
				RecvType: recv,
				Args:     args,
			})
		}
	}

	var g glue
	g.Package = "builtin"
	g.Types = export

	t, err := template.New("code").Parse(codeTemplate2)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(os.Stdout, &g)
	if err != nil {
		log.Fatal(err)
	}
}
