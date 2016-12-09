package parser

import (
	"io"
	"os"
	"strings"
)

type markingReader struct {
	r   *strings.Reader
	err error
	pos int

	furthest int64
}

func (m *markingReader) Mark() int {
	pos, _ := m.r.Seek(0, os.SEEK_CUR)
	if pos > m.furthest {
		m.furthest = pos
	}

	return int(pos)
}

func (m *markingReader) Rewind(p int) {
	m.r.Seek(int64(p), os.SEEK_SET)
}

func (m *markingReader) RuneScanner() io.RuneScanner {
	return m.r
}
