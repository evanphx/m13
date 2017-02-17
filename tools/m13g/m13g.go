package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/parser"
)

var fFile = flag.String("file", "", "path to m13g file")

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(*fFile)
	if err != nil {
		log.Fatal(err)
	}

	parser, err := parser.NewParser(string(data))
	if err != nil {
		log.Fatal(err)
	}

	node, err := parser.ParseG()
	if err != nil {
		log.Fatal(err)
	}

	// gast.Print(nil, node)

	gen(node)
}

var fileTemp = `
package {{.Package}}

{{range .Imports}}
	import "{{.}}"
{{end}}

{{range .Types}}
	{{.}}
{{end}}
`

type fileInfo struct {
	Package string
	Imports []string
	Types   []string
}

func gen(top ast.Node) {
	blk, ok := top.(*ast.Block)
	if !ok {
		return
	}

	var info fileInfo

	for _, expr := range blk.Expressions {
		switch st := expr.(type) {
		case *ast.Package:
			info.Package = st.Name
		case *ast.Import:
			info.Imports = append(info.Imports, strings.Join(st.Path, "/"))
		case *ast.ClassDefinition:
			info.Types = append(info.Types, genClass(st))
		}
	}

	tmpl, err := template.New("file").Parse(fileTemp)
	if err != nil {
		panic(err)
	}

	tmpl.Execute(os.Stdout, &info)
}

func cleanName(name string) string {
	switch name {
	case "<<":
		return "append_op"
	default:
		return name
	}
}

type memberInfo struct {
	Name, Type string
}

type argumentInfo struct {
	Name, Type string
}

type methodInfo struct {
	Name      string
	CleanName string
	Arguments []argumentInfo
	GoCode    string
}

type classInfo struct {
	Name    string
	Members []memberInfo
	Methods []methodInfo
}

var typeTemp = `
{{$name := .Name}}
type {{.Name}} struct {
	{{range .Members}}
		Object

		{{.Name}} {{.Type}}
	{{end}}
}

{{range .Methods}}

func meth{{$name}}{{.CleanName}}(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
	self := recv.(*{{$name}})

	{{range $index, $arg := .Arguments}}
		{{.Name}} := args[{{$index}}].({{.Type}})
	{{end}}

	{{.GoCode}}

	return recv, nil
}
{{end}}

func init{{.Name}}(pkg *Package, cls *Class) {
	{{range .Methods}}
		cls.AddMethod(&MethodDescriptor{
			Name: "{{.Name}}",
			Signature: Signature{
				Required: {{len .Arguments}},
			},
			Func: meth{{$name}}{{.CleanName}},
		})
	{{end}}
}
`

func genClass(cls *ast.ClassDefinition) string {
	body, ok := cls.Body.(*ast.Block)
	if !ok {
		return ""
	}

	var (
		members []memberInfo
		methods []methodInfo
	)

	for _, expr := range body.Expressions {
		switch st := expr.(type) {
		case *ast.Has:
			if st.Type == nil {
				st.Type = &ast.Type{Name: "Value"}
			}

			members = append(members, memberInfo{
				Name: st.Variable,
				Type: st.Type.Name,
			})
		case *ast.GoDefinition:
			var args []argumentInfo

			for _, arg := range st.Arguments {
				if arg.Type == nil {
					arg.Type = &ast.Type{Name: "Value"}
				}

				args = append(args, argumentInfo{
					Name: arg.Name,
					Type: arg.Type.Name,
				})
			}

			methods = append(methods, methodInfo{
				Name:      st.Name,
				CleanName: cleanName(st.Name),
				Arguments: args,
				GoCode:    st.Body,
			})
		}
	}

	info := classInfo{
		Name:    cls.Name,
		Members: members,
		Methods: methods,
	}

	tmpl, err := template.New("class").Parse(typeTemp)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, &info)
	if err != nil {
		panic(err)
	}

	return buf.String()
}
