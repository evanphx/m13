package parser

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/evanphx/m13/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestParser(t *testing.T) {
	n := neko.Start(t)

	n.It("parses an Integer", func() {
		parser, err := NewParser("10")

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Integer)
		require.True(t, ok)

		assert.Equal(t, int64(10), n.Value)
	})

	n.It("parses a String", func() {
		src := `"hello"`

		parser, err := NewParser(src)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.String)
		require.True(t, ok)

		assert.Equal(t, "hello", n.Value)
	})

	n.It("parses an Atom", func() {
		src := `:foo`

		parser, err := NewParser(src)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Atom)
		require.True(t, ok)

		assert.Equal(t, "foo", n.Value)
	})

	n.It("parses a True", func() {
		src := `true`

		parser, err := NewParser(src)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		_, ok := tree.(*ast.True)
		require.True(t, ok)
	})

	n.It("parses a False", func() {
		src := `false`

		parser, err := NewParser(src)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		_, ok := tree.(*ast.False)
		assert.True(t, ok)
	})

	n.It("parses a Nil", func() {
		src := `nil`

		parser, err := NewParser(src)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		_, ok := tree.(*ast.Nil)
		assert.True(t, ok)
	})

	n.It("parses a variable", func() {
		src := `a`

		parser, err := NewParser(src)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		n, ok := tree.(*ast.Variable)
		require.True(t, ok)

		assert.Equal(t, "a", n.Name)
	})

	n.It("parses a method call", func() {
		src := `a.b()`

		parser, err := NewParser(src)
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
		src := `a.b(c)`

		parser, err := NewParser(src)
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
		src := `a.b(c,d,e,f)`

		parser, err := NewParser(src)
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
		src := `a.b().c()`

		parser, err := NewParser(src)
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
		src := `a.b().c(d,e)`

		parser, err := NewParser(src)
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
		src := "a=1"

		parser, err := NewParser(src)
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
		src := "=>1"

		parser, err := NewParser(src)
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
		src := "x=>1"

		parser, err := NewParser(src)
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
		src := "(x, y) => 1"

		parser, err := NewParser(src)
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
		src := "x => { 1 }"

		parser, err := NewParser(src)
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
		src := "x => { 1; 2 }"

		parser, err := NewParser(src)
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
		src := "x => { 1\n 2 }"

		parser, err := NewParser(src)
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
		src := "x => {\n  1\n}"

		parser, err := NewParser(src)
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
		src := "1\n2"

		parser, err := NewParser(src)
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
		src := "1\n\n2"

		parser, err := NewParser(src)
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
		src := "import a.b.c"

		parser, err := NewParser(src)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		i, ok := tree.(*ast.Import)
		require.True(t, ok)

		assert.Equal(t, []string{"a", "b", "c"}, i.Path)
	})

	n.It("parses a test program", func() {
		src := `import os; os.stdout().puts("hello m13")`

		parser, err := NewParser(src)
		require.NoError(t, err)

		_, err = parser.Parse()
		require.NoError(t, err)
	})

	n.It("parses a test program with spaces", func() {
		src := `
		import os;

os.stdout().puts("hello m13");`

		parser, err := NewParser(src)
		require.NoError(t, err)

		_, err = parser.Parse()
		require.NoError(t, err)
	})

	n.It("parses a method definition with no args", func() {
		src := `def foo { 1 }`

		parser, err := NewParser(src)
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
		src := `def foo(a,b) { 1 }`

		parser, err := NewParser(src)
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
		src := `def foo(a) { 1 }`

		parser, err := NewParser(src)
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
		src := `class Blah { 1 }`

		parser, err := NewParser(src)
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
		src := `# hello, newman`

		parser, err := NewParser(src)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		def, ok := tree.(*ast.Comment)
		require.True(t, ok)

		assert.Equal(t, " hello, newman", def.Comment)
	})

	n.It("parser an ivar", func() {
		src := `@age`

		parser, err := NewParser(src)
		require.NoError(t, err)

		tree, err := parser.Parse()
		require.NoError(t, err)

		def, ok := tree.(*ast.IVar)
		require.True(t, ok)

		assert.Equal(t, "age", def.Name)
	})

	n.It("parser a class definition with ivar decls", func() {
		src := `class Blah { has @age }`

		parser, err := NewParser(src)
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
		src := `class Blah { has @age is rw }`

		parser, err := NewParser(src)
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
		src := `class Blah { has @age is rw is locked}`

		parser, err := NewParser(src)
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
		src := `3 + 4`

		parser, err := NewParser(src)
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
		src := `3 + 4 * 2`

		parser, err := NewParser(src)
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
		src := `3 * 4 + 2`

		parser, err := NewParser(src)
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
		src := `3 * 4 + 2 * 5`

		parser, err := NewParser(src)
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
		src := `3 ** 4 + 2 ** 5`

		parser, err := NewParser(src)
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
		src := `if a { b }`

		parser, err := NewParser(src)
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
		src := `while a { b }`

		parser, err := NewParser(src)
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
		src := `a++`

		parser, err := NewParser(src)
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
		src := `a--`

		parser, err := NewParser(src)
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
		src := `a(1)`

		parser, err := NewParser(src)
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

		parser, err := NewParser(s)
		require.NoError(t, err, s)

		tree, err := parser.Parse()
		require.NoError(t, err, s)

		t.Logf("=> %#v", tree)
	}
}

func TestMethodParses(t *testing.T) {
	n := neko.Start(t)

	n.It("parses an attribute access", func() {
		src := `a.b`

		parser, err := NewParser(src)
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
		src := `1.b`

		parser, err := NewParser(src)
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
		src := `1.class`

		parser, err := NewParser(src)
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
		src := `a.b.c`

		parser, err := NewParser(src)
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
		src := `a.b.c().d.e().f`

		parser, err := NewParser(src)
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
		src := `a.c().b`

		parser, err := NewParser(src)
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
		src := `a.b = 3`

		parser, err := NewParser(src)
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
		src := `a.c().b = 3`

		parser, err := NewParser(src)
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
		src := `a.^b()`

		parser, err := NewParser(src)
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
		src := `a.b 3`

		parser, err := NewParser(src)
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

	n.It("parses a method call without parens and a lambda", func() {
		src := `a.b "d", x => { 3 }`

		parser, err := NewParser(src)
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

	parser, err := NewParser(string(data))
	require.NoError(t, err)

	_, err = parser.Parse()
	require.NoError(t, err, string(data))
}
