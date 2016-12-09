package parser

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/lex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestParser(t *testing.T) {
	n := neko.Start(t)

	n.It("parses an Integer", func() {
		lex, err := lex.NewLexer("10")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(10), n.Value)
	})

	n.It("parses a String", func() {
		lex, err := lex.NewLexer(`"hello"`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.String)
		require.True(t, ok)

		assert.Equal(t, "hello", n.Value)
	})

	n.It("parses an Atom", func() {
		lex, err := lex.NewLexer(`:foo`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Atom)
		require.True(t, ok)

		assert.Equal(t, "foo", n.Value)
	})

	n.It("parses a True", func() {
		lex, err := lex.NewLexer(`true`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		_, ok := tree.(*ast.True)
		require.True(t, ok)
	})

	n.It("parses a False", func() {
		lex, err := lex.NewLexer(`false`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		_, ok := tree.(*ast.False)
		assert.True(t, ok)
	})

	n.It("parses a Nil", func() {
		lex, err := lex.NewLexer(`nil`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		_, ok := tree.(*ast.Nil)
		assert.True(t, ok)
	})

	n.It("parses a variable", func() {
		lex, err := lex.NewLexer(`a`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", n.Name)
	})

	n.It("parses a method call", func() {
		lex, err := lex.NewLexer(`a.b()`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Call)
		require.True(t, ok)

		c, ok := n.Receiver.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", c.Name)

		assert.Equal(t, "b", n.MethodName)
	})

	n.It("parses a method call with an arg", func() {
		lex, err := lex.NewLexer(`a.b(c)`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Call)
		require.True(t, ok)

		c, ok := n.Receiver.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", c.Name)

		assert.Equal(t, "b", n.MethodName)

		require.Equal(t, 1, len(n.Args))

		a, ok := n.Args[0].(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "c", a.Name)
	})

	n.It("parses a method call with args", func() {
		lex, err := lex.NewLexer(`a.b(c,d,e,f)`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Call)
		require.True(t, ok)

		c, ok := n.Receiver.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", c.Name)

		assert.Equal(t, "b", n.MethodName)

		require.Equal(t, 4, len(n.Args))

		x, ok := n.Args[0].(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "c", x.Name)

		d, ok := n.Args[1].(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "d", d.Name)

		d, ok = n.Args[2].(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "e", d.Name)

		d, ok = n.Args[3].(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "f", d.Name)
	})

	n.It("parses a chained method call", func() {
		lex, err := lex.NewLexer(`a.b().c()`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Call)
		require.True(t, ok)

		c, ok := n.Receiver.(*ast.Call)
		require.True(t, ok)

		assert.Equal(t, "b", c.MethodName)

		assert.Equal(t, "c", n.MethodName)
	})

	n.It("parses a chained method call with args", func() {
		lex, err := lex.NewLexer(`a.b().c(d,e)`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Call)
		require.True(t, ok)

		c, ok := n.Receiver.(*ast.Call)
		require.True(t, ok)

		assert.Equal(t, "b", c.MethodName)

		assert.Equal(t, "c", n.MethodName)

		require.Equal(t, 2, len(n.Args))
	})

	n.It("parser a local variable assignment", func() {
		lex, err := lex.NewLexer("a=1")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Assign)
		require.True(t, ok)

		assert.Equal(t, "a", n.Name)

		v, ok := n.Value.(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), v.Value)
	})

	n.It("parses a simple lambda", func() {
		lex, err := lex.NewLexer("=>1")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Lambda)
		require.True(t, ok)

		assert.Equal(t, 0, len(n.Args))

		v, ok := n.Expr.(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), v.Value)
	})

	n.It("parses a lambda with arg", func() {
		lex, err := lex.NewLexer("x=>1")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Lambda)
		require.True(t, ok)

		assert.Equal(t, 1, len(n.Args))

		assert.Equal(t, "x", n.Args[0])

		v, ok := n.Expr.(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), v.Value)
	})

	n.It("parses a lambda with args", func() {
		lex, err := lex.NewLexer("(x, y) => 1")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Lambda)
		require.True(t, ok)

		assert.Equal(t, 2, len(n.Args))

		assert.Equal(t, "x", n.Args[0])
		assert.Equal(t, "y", n.Args[1])

		v, ok := n.Expr.(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), v.Value)

	})

	n.It("parses a lambda with a brace body", func() {
		lex, err := lex.NewLexer("x => { 1 }")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Lambda)
		require.True(t, ok)

		assert.Equal(t, 1, len(n.Args))

		assert.Equal(t, "x", n.Args[0])

		fmt.Printf("lambda: %#v\n", n)

		b, ok := n.Expr.(*ast.Block)
		require.True(t, ok, fmt.Sprintf("%T", n.Expr))

		v, ok := b.Expressions[0].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), v.Value)
	})

	n.It("parses a lambda with multiple expressions", func() {
		lex, err := lex.NewLexer("x => { 1; 2 }")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		if err != nil {
			fmt.Printf(err.Error())
		}
		require.NoError(t, err)

		n, ok := tree.(*ast.Lambda)
		require.True(t, ok)

		assert.Equal(t, 1, len(n.Args))

		assert.Equal(t, "x", n.Args[0])

		b, ok := n.Expr.(*ast.Block)
		require.True(t, ok)

		v, ok := b.Expressions[0].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), v.Value)

		v, ok = b.Expressions[1].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(2), v.Value)

	})

	n.It("parses a lambda with multiple expressions using a newline seperator", func() {
		lex, err := lex.NewLexer("x => { 1\n 2 }")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Lambda)
		require.True(t, ok)

		assert.Equal(t, 1, len(n.Args))

		assert.Equal(t, "x", n.Args[0])

		b, ok := n.Expr.(*ast.Block)
		require.True(t, ok)

		v, ok := b.Expressions[0].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), v.Value)

		v, ok = b.Expressions[1].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(2), v.Value)
	})

	n.It("parses a lambda with on it's only line", func() {
		lex, err := lex.NewLexer("x => {\n  1\n}")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Lambda)
		require.True(t, ok)

		assert.Equal(t, 1, len(n.Args))

		assert.Equal(t, "x", n.Args[0])

		b, ok := n.Expr.(*ast.Block)
		require.True(t, ok)

		v, ok := b.Expressions[0].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), v.Value)
	})

	n.It("parses a toplevel expression sequence", func() {
		lex, err := lex.NewLexer("1\n2")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		b, ok := tree.(*ast.Block)
		require.True(t, ok)

		v, ok := b.Expressions[0].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), v.Value)

		v, ok = b.Expressions[1].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(2), v.Value)
	})

	n.It("allows blank lines between statements", func() {
		lex, err := lex.NewLexer("1\n\n2")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		b, ok := tree.(*ast.Block)
		require.True(t, ok)

		v, ok := b.Expressions[0].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), v.Value)

		v, ok = b.Expressions[1].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(2), v.Value)
	})

	n.It("parses an import", func() {
		lex, err := lex.NewLexer("import a.b.c")
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		i, ok := tree.(*ast.Import)
		require.True(t, ok)

		assert.Equal(t, []string{"a", "b", "c"}, i.Path)
	})

	n.It("parses a test program", func() {
		prog := `import os; os.stdout().puts("hello m13")`

		lex, err := lex.NewLexer(prog)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		_, err = parser.Parse()
		require.NoError(t, err)
	})

	n.It("parses a test program with spaces", func() {
		prog := `
		import os;

os.stdout().puts("hello m13");`

		lex, err := lex.NewLexer(prog)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		_, err = parser.Parse()
		require.NoError(t, err)
	})

	n.It("parses a method definition with no args", func() {
		lex, err := lex.NewLexer(`def foo { 1 }`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		def, ok := tree.(*ast.Definition)
		require.True(t, ok)

		assert.Equal(t, "foo", def.Name)

		assert.Equal(t, 0, len(def.Arguments))

		blk, ok := def.Body.(*ast.Block)
		require.True(t, ok)

		assert.Equal(t, int64(1), blk.Expressions[0].(*ast.Integer).Value)
	})

	n.It("parses a method definition with 2 args", func() {
		lex, err := lex.NewLexer(`def foo(a,b) { 1 }`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		def, ok := tree.(*ast.Definition)
		require.True(t, ok)

		assert.Equal(t, "foo", def.Name)

		assert.Equal(t, []string{"a", "b"}, def.Arguments)

		blk, ok := def.Body.(*ast.Block)
		require.True(t, ok)

		assert.Equal(t, int64(1), blk.Expressions[0].(*ast.Integer).Value)
	})

	n.It("parses a method definition with 1 arg", func() {
		lex, err := lex.NewLexer(`def foo(a) { 1 }`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		def, ok := tree.(*ast.Definition)
		require.True(t, ok)

		assert.Equal(t, "foo", def.Name)

		assert.Equal(t, []string{"a"}, def.Arguments)

		blk, ok := def.Body.(*ast.Block)
		require.True(t, ok)

		assert.Equal(t, int64(1), blk.Expressions[0].(*ast.Integer).Value)
	})

	n.It("parser a class definition", func() {
		lex, err := lex.NewLexer(`class Blah { 1 }`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		def, ok := tree.(*ast.ClassDefinition)
		require.True(t, ok)

		assert.Equal(t, "Blah", def.Name)

		blk, ok := def.Body.(*ast.Block)
		require.True(t, ok)

		assert.Equal(t, int64(1), blk.Expressions[0].(*ast.Integer).Value)
	})

	n.It("parses a comment", func() {
		lex, err := lex.NewLexer(`# hello, newman`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		def, ok := tree.(*ast.Comment)
		require.True(t, ok)

		assert.Equal(t, " hello, newman", def.Comment)
	})

	n.It("parser an ivar", func() {
		lex, err := lex.NewLexer(`@age`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		def, ok := tree.(*ast.IVar)
		require.True(t, ok)

		assert.Equal(t, "age", def.Name)
	})

	n.It("parser a class definition with ivar decls", func() {
		lex, err := lex.NewLexer(`class Blah { has @age }`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		def, ok := tree.(*ast.ClassDefinition)
		require.True(t, ok)

		assert.Equal(t, "Blah", def.Name)

		blk, ok := def.Body.(*ast.Block)
		require.True(t, ok)

		has, ok := blk.Expressions[0].(*ast.Has)
		require.True(t, ok)

		assert.Equal(t, "age", has.Variable)
	})

	n.It("parser a class definition with ivar decls and trait", func() {
		lex, err := lex.NewLexer(`class Blah { has @age is rw }`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		def, ok := tree.(*ast.ClassDefinition)
		require.True(t, ok)

		assert.Equal(t, "Blah", def.Name)

		blk, ok := def.Body.(*ast.Block)
		require.True(t, ok)

		has, ok := blk.Expressions[0].(*ast.Has)
		require.True(t, ok)

		assert.Equal(t, "age", has.Variable)

		assert.Equal(t, []string{"rw"}, has.Traits)
	})

	n.It("parser a class definition with ivar decls and traits", func() {
		lex, err := lex.NewLexer(`class Blah { has @age is rw is locked}`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		def, ok := tree.(*ast.ClassDefinition)
		require.True(t, ok)

		assert.Equal(t, "Blah", def.Name)

		blk, ok := def.Body.(*ast.Block)
		require.True(t, ok)

		has, ok := blk.Expressions[0].(*ast.Has)
		require.True(t, ok)

		assert.Equal(t, "age", has.Variable)

		assert.Equal(t, []string{"rw", "locked"}, has.Traits)
	})

	n.It("parses `3 + 4`", func() {
		lex, err := lex.NewLexer(`3 + 4`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		op, ok := tree.(*ast.Op)
		require.True(t, ok)

		assert.Equal(t, "+", op.Name)

		assert.Equal(t, int64(3), op.Left.(*ast.Integer).Value)
		assert.Equal(t, int64(4), op.Right.(*ast.Integer).Value)
	})

	n.It("parses `3 + 4 * 2`", func() {
		lex, err := lex.NewLexer(`3 + 4 * 2`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		op, ok := tree.(*ast.Op)
		require.True(t, ok)

		assert.Equal(t, "+", op.Name)

		assert.Equal(t, int64(3), op.Left.(*ast.Integer).Value)

		op2, ok := op.Right.(*ast.Op)
		require.True(t, ok)

		assert.Equal(t, "*", op2.Name)

		assert.Equal(t, int64(4), op2.Left.(*ast.Integer).Value)
		assert.Equal(t, int64(2), op2.Right.(*ast.Integer).Value)
	})

	n.It("parses `3 * 4 + 2`", func() {
		lex, err := lex.NewLexer(`3 * 4 + 2`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		op, ok := tree.(*ast.Op)
		require.True(t, ok)

		assert.Equal(t, "+", op.Name)

		assert.Equal(t, int64(2), op.Right.(*ast.Integer).Value)

		op2, ok := op.Left.(*ast.Op)
		require.True(t, ok)

		assert.Equal(t, "*", op2.Name)

		assert.Equal(t, int64(3), op2.Left.(*ast.Integer).Value)
		assert.Equal(t, int64(4), op2.Right.(*ast.Integer).Value)
	})

	n.It("parses `3 * 4 + 2 * 5`", func() {
		lex, err := lex.NewLexer(`3 * 4 + 2 * 5`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		op, ok := tree.(*ast.Op)
		require.True(t, ok)

		assert.Equal(t, "+", op.Name)

		op2, ok := op.Left.(*ast.Op)
		require.True(t, ok)

		assert.Equal(t, "*", op2.Name)

		assert.Equal(t, int64(3), op2.Left.(*ast.Integer).Value)
		assert.Equal(t, int64(4), op2.Right.(*ast.Integer).Value)

		op3, ok := op.Right.(*ast.Op)
		require.True(t, ok, "%T", op.Right)

		assert.Equal(t, "*", op3.Name)

		assert.Equal(t, int64(2), op3.Left.(*ast.Integer).Value)
		assert.Equal(t, int64(5), op3.Right.(*ast.Integer).Value)
	})

	n.It("parses `3 ** 4 + 2 ** 5`", func() {
		lex, err := lex.NewLexer(`3 ** 4 + 2 ** 5`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		op, ok := tree.(*ast.Op)
		require.True(t, ok)

		assert.Equal(t, "+", op.Name)

		op2, ok := op.Left.(*ast.Op)
		require.True(t, ok)

		assert.Equal(t, "**", op2.Name)

		assert.Equal(t, int64(3), op2.Left.(*ast.Integer).Value)
		assert.Equal(t, int64(4), op2.Right.(*ast.Integer).Value)

		op3, ok := op.Right.(*ast.Op)
		require.True(t, ok, "%T", op.Right)

		assert.Equal(t, "**", op3.Name)

		assert.Equal(t, int64(2), op3.Left.(*ast.Integer).Value)
		assert.Equal(t, int64(5), op3.Right.(*ast.Integer).Value)
	})

	n.It("parses an if", func() {
		lex, err := lex.NewLexer(`if a { b }`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		ift, ok := tree.(*ast.If)
		require.True(t, ok)

		cond, ok := ift.Cond.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", cond.Name)

		body, ok := ift.Body.(*ast.Block)
		require.True(t, ok)

		v, ok := body.Expressions[0].(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "b", v.Name)
	})

	n.It("parses a while", func() {
		lex, err := lex.NewLexer(`while a { b }`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		ift, ok := tree.(*ast.While)
		require.True(t, ok)

		cond, ok := ift.Cond.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", cond.Name)

		body, ok := ift.Body.(*ast.Block)
		require.True(t, ok)

		v, ok := body.Expressions[0].(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "b", v.Name)
	})

	n.It("parses a++", func() {
		lex, err := lex.NewLexer(`a++`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		inc, ok := tree.(*ast.Inc)
		require.True(t, ok)

		v, ok := inc.Receiver.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", v.Name)
	})

	n.It("parses a--", func() {
		lex, err := lex.NewLexer(`a--`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		inc, ok := tree.(*ast.Dec)
		require.True(t, ok)

		v, ok := inc.Receiver.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", v.Name)
	})

	n.It("parses a function invoke", func() {
		lex, err := lex.NewLexer(`a(1)`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		inv, ok := tree.(*ast.Invoke)
		require.True(t, ok)

		assert.Equal(t, "a", inv.Name)

		require.Equal(t, 1, len(inv.Args))

		lit, ok := inv.Args[0].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), lit.Value)
	})

	n.Meow()
}

func TestRandomSnippits(t *testing.T) {
	var snippits = []string{
		`a(4)`,
		`a = x => x + 3`,
		`a = x => x + 3; a(4)`,
		`a.^b`,
		`a.^b()`,
		`3.^class()`,
		`3.^class`,
		`3.^class.name`,
		`c.expect(3.^class.name)`,
		`3 == 4`,
		`a.b == c.d`,
		`c.expect(3.^class.name) == "builtin.I64"`,
		`a.b c, d`,
	}

	for _, s := range snippits {
		t.Logf("parsing: %s", s)
		lex, err := lex.NewLexer(s)
		require.NoError(t, err, s)

		parser, err := NewParser(lex)
		require.NoError(t, err, s)

		tree, err := parser.Parse()
		require.NoError(t, err, s)

		t.Logf("=> %#v", tree)
	}
}

func TestMethodParses(t *testing.T) {
	n := neko.Start(t)

	n.It("parses an attribute access", func() {
		lex, err := lex.NewLexer(`a.b`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		attr, ok := tree.(*ast.Attribute)
		require.True(t, ok)

		assert.Equal(t, "b", attr.Name)

		obj, ok := attr.Receiver.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", obj.Name)
	})

	n.It("parses an attribute access off a number", func() {
		lex, err := lex.NewLexer(`1.b`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		attr, ok := tree.(*ast.Attribute)
		require.True(t, ok)

		assert.Equal(t, "b", attr.Name)

		obj, ok := attr.Receiver.(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), obj.Value)
	})

	n.It("allows keywords in attribute names", func() {
		lex, err := lex.NewLexer(`1.class`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		attr, ok := tree.(*ast.Attribute)
		require.True(t, ok)

		assert.Equal(t, "class", attr.Name)

		obj, ok := attr.Receiver.(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(1), obj.Value)
	})

	n.It("parses a nested attribute access", func() {
		lex, err := lex.NewLexer(`a.b.c`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		attr, ok := tree.(*ast.Attribute)
		require.True(t, ok)

		assert.Equal(t, "c", attr.Name)

		attr2, ok := attr.Receiver.(*ast.Attribute)
		require.True(t, ok)

		assert.Equal(t, "b", attr2.Name)

		obj, ok := attr2.Receiver.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", obj.Name)
	})

	n.It("parses a gnarly attr+call chain", func() {
		lex, err := lex.NewLexer(`a.b.c().d.e().f`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		attr, ok := tree.(*ast.Attribute)
		require.True(t, ok)

		assert.Equal(t, "f", attr.Name)

		call, ok := attr.Receiver.(*ast.Call)
		require.True(t, ok)

		assert.Equal(t, "e", call.MethodName)
	})

	n.It("parses an attribute access off a call", func() {
		lex, err := lex.NewLexer(`a.c().b`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		attr, ok := tree.(*ast.Attribute)
		require.True(t, ok)

		assert.Equal(t, "b", attr.Name)

		obj, ok := attr.Receiver.(*ast.Call)
		require.True(t, ok)

		assert.Equal(t, "c", obj.MethodName)
	})

	n.It("parses an attribute assign", func() {
		lex, err := lex.NewLexer(`a.b = 3`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		attr, ok := tree.(*ast.AttributeAssign)
		require.True(t, ok)

		assert.Equal(t, "b", attr.Name)

		obj, ok := attr.Receiver.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", obj.Name)

		val, ok := attr.Value.(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(3), val.Value)
	})

	n.It("parses an attribute assign off a call", func() {
		lex, err := lex.NewLexer(`a.c().b = 3`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		attr, ok := tree.(*ast.AttributeAssign)
		require.True(t, ok)

		assert.Equal(t, "b", attr.Name)

		obj, ok := attr.Receiver.(*ast.Call)
		require.True(t, ok)

		assert.Equal(t, "c", obj.MethodName)

		val, ok := attr.Value.(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(3), val.Value)
	})

	n.It("parses an up method call", func() {
		lex, err := lex.NewLexer(`a.^b()`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.UpCall)
		require.True(t, ok)

		c, ok := n.Receiver.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", c.Name)

		assert.Equal(t, "b", n.MethodName)
	})

	n.It("parses a method call without parens", func() {
		lex, err := lex.NewLexer(`a.b 3`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Call)
		require.True(t, ok)

		c, ok := n.Receiver.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", c.Name)

		assert.Equal(t, "b", n.MethodName)

		require.Equal(t, 1, len(n.Args))

		i, ok := n.Args[0].(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(3), i.Value)
	})

	n.Only("parses a method call without parens and a lambda", func() {
		lex, err := lex.NewLexer(`a.b "d", x => { 3 }`)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Call)
		require.True(t, ok)

		fmt.Printf("call: %#v\n", n)

		c, ok := n.Receiver.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", c.Name)

		assert.Equal(t, "b", n.MethodName)

		require.Equal(t, 2, len(n.Args))

		i, ok := n.Args[0].(*ast.String)
		require.True(t, ok)

		assert.Equal(t, "d", i.Value)

		l, ok := n.Args[1].(*ast.Lambda)
		require.True(t, ok)

		require.Equal(t, 1, len(l.Args))

		assert.Equal(t, "x", l.Args[0])
	})

	n.Meow()
}

func TestBasic(t *testing.T) {
	data, err := ioutil.ReadFile("../test/basic.m13")
	require.NoError(t, err)

	lex, err := lex.NewLexer(string(data))
	require.NoError(t, err)

	parser, err := NewParser(lex)
	require.NoError(t, err)

	_, err = parser.Parse()
	require.NoError(t, err, string(data))
}
