package lex

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestLexer(t *testing.T) {
	var intTests = []struct {
		s string
		i int64
	}{
		{"47", 47},
		{"0x47", 0x47},
		{"0", 0},
		{"1", 1},
	}

	for _, it := range intTests {
		lex, err := NewLexer(it.s)
		require.NoError(t, err)

		x, err := lex.Next()
		require.NoError(t, err)

		assert.Equal(t, Integer, x.Type)
		assert.Equal(t, it.i, x.Value.(int64))
	}

	var strTests = []struct {
		i string
		o string
	}{
		{`""`, ""},
		{`"hello"`, "hello"},
		{`"hello\n"`, "hello\n"},
		{`"hello\r"`, "hello\r"},
		{`"hello\t"`, "hello\t"},
		{`"got \u0021"`, "got \u0021"},
		{`"got \U00000021"`, "got \U00000021"},
	}

	for _, st := range strTests {
		lex, err := NewLexer(st.i)
		require.NoError(t, err)

		x, err := lex.Next()
		require.NoError(t, err)

		assert.Equal(t, String, x.Type)
		assert.Equal(t, st.o, x.Value.(string))
	}

	var atomTests = []struct {
		i string
		o string
	}{
		{":hello", "hello"},
		{":123", "123"},
		{":a1", "a1"},
	}

	for _, st := range atomTests {
		lex, err := NewLexer(st.i)
		require.NoError(t, err)

		x, err := lex.Next()
		require.NoError(t, err)

		assert.Equal(t, Atom, x.Type)
		assert.Equal(t, st.o, x.Value.(string))
	}

	var keywordTests = []struct {
		i string
		o Type
	}{
		{"true", True},
		{"false", False},
		{"nil", Nil},
		{".", Dot},
		{".^", UpDot},
		{"(", OpenParen},
		{")", CloseParen},
		{",", Comma},
		{"=", Equal},
		{"=>", Into},
		{"{", OpenBrace},
		{"}", CloseBrace},
		{";", Semi},
		{"\n", Newline},
		{"import", Import},
		{"def", Def},
		{"class", Class},
		{"has", Has},
		{"is", Is},
		{"if", If},
		{"++", Inc},
		{"--", Dec},
		{"while", While},
	}

	for _, st := range keywordTests {
		lex, err := NewLexer(st.i)
		require.NoError(t, err)

		x, err := lex.Next()
		require.NoError(t, err, "parsing: %s", st.i)

		assert.Equal(t, st.o, x.Type, x.Type.String())
	}

	var wordTests = []string{
		"foo",
		"nik",
		"tri",
		"fab",
	}

	for _, v := range wordTests {
		lex, err := NewLexer(v)
		require.NoError(t, err)

		x, err := lex.Next()
		require.NoError(t, err)

		assert.Equal(t, Word, x.Type)
		assert.Equal(t, v, x.Value.(string))
	}

	var ivarTests = []string{
		"@foo",
	}

	for _, v := range ivarTests {
		lex, err := NewLexer(v)
		require.NoError(t, err)

		x, err := lex.Next()
		require.NoError(t, err)

		assert.Equal(t, IVar, x.Type)
		assert.Equal(t, v[1:], x.Value.(string))
	}

	var opTests = []string{
		"+",
		">",
		"<",
		"<+>",
		"==",
	}

	for _, v := range opTests {
		lex, err := NewLexer(v)
		require.NoError(t, err)

		x, err := lex.Next()
		require.NoError(t, err)

		require.Equal(t, Operator, x.Type, x.Type.String())
		assert.Equal(t, v, x.Value.(string))
	}

	lex, err := NewLexer("true")
	require.NoError(t, err)

	_, err = lex.Next()
	require.NoError(t, err)

	n, err := lex.Next()
	require.NoError(t, err)

	assert.Equal(t, Term, n.Type)
}

func TestLexerEdge(t *testing.T) {
	n := neko.Start(t)

	n.It("handles a=1", func() {
		lex, err := NewLexer("a=1")
		require.NoError(t, err)

		v, err := lex.Next()
		require.NoError(t, err)

		assert.Equal(t, Word, v.Type)

		v, err = lex.Next()
		require.NoError(t, err)

		assert.Equal(t, Equal, v.Type)

		v, err = lex.Next()
		require.NoError(t, err)

		assert.Equal(t, Integer, v.Type)
	})

	n.It("ignores space", func() {
		lex, err := NewLexer(" a   b   ")
		require.NoError(t, err)

		v, err := lex.Next()
		require.NoError(t, err)

		assert.Equal(t, Word, v.Type)
		assert.Equal(t, "a", v.Value.(string))

		v, err = lex.Next()
		require.NoError(t, err)

		assert.Equal(t, Word, v.Type)
		assert.Equal(t, "b", v.Value.(string))
	})

	n.It("ignores space", func() {
		lex, err := NewLexer(" # blah")
		require.NoError(t, err)

		v, err := lex.Next()
		require.NoError(t, err)

		assert.Equal(t, Comment, v.Type)
		assert.Equal(t, " blah", v.Value.(string))
	})

	n.It("has a++ as 2 lexems", func() {
		lex, err := NewLexer("a++")
		require.NoError(t, err)

		v, err := lex.Next()
		require.NoError(t, err)

		assert.Equal(t, Word, v.Type)
		assert.Equal(t, "a", v.Value.(string))

		v, err = lex.Next()
		require.NoError(t, err)

		assert.Equal(t, Inc, v.Type)
	})

	n.Meow()
}
