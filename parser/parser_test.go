package parser

import (
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

		b, ok := n.Expr.(*ast.Block)
		require.True(t, ok)

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

	n.It("parses a test program", func() {
		prog := `import os; os.stdout().puts("hello m13")`

		lex, err := lex.NewLexer(prog)
		require.NoError(t, err)

		parser, err := NewParser(lex)
		require.NoError(t, err)

		_, err = parser.Parse()
		require.NoError(t, err)
	})

	n.Meow()
}
