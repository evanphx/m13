package parser

import (
	"io"
	"os"
	"strings"

	"github.com/evanphx/m13/lex"
)

type markingLexer struct {
	r   *strings.Reader
	err error
	pos int

	furthest int64
}

func (m *markingLexer) Next() *lex.Value {
	panic("nope")
	return nil
}

func (m *markingLexer) Mark() int {
	pos, _ := m.r.Seek(0, os.SEEK_CUR)
	if pos > m.furthest {
		m.furthest = pos
	}

	return int(pos)
}

func (m *markingLexer) Rewind(p int) {
	m.r.Seek(int64(p), os.SEEK_SET)
}

func (m *markingLexer) RuneScanner() io.RuneScanner {
	return m.r
}
