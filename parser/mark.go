package parser

import "github.com/evanphx/m13/lex"

type markingLexer struct {
	lex *lex.Lexer
	err error
	pos int

	ready []*lex.Value
}

func (m *markingLexer) Next() *lex.Value {
	if m.pos < len(m.ready) {
		v := m.ready[m.pos]
		m.pos++

		return v
	}

	v, err := m.lex.Next()
	if err != nil {
		m.err = err
		return nil
	}

	// fmt.Printf("=> %s\n", v.Type)

	m.pos++

	m.ready = append(m.ready, v)

	return v
}

func (m *markingLexer) Mark() int {
	return m.pos
}

func (m *markingLexer) Rewind(p int) {
	m.pos = p
}
