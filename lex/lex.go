package lex

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

type Reader interface {
	io.RuneScanner
	io.ReaderAt
}

type Lexer struct {
	Source string
	r      Reader
	pos    int
}

//go:generate stringer -type=Type
type Type int

const (
	Unknown Type = iota
	Term
	Integer
	String
	Atom
	True
	False
	Nil
	Dot
	Word
	Operator
	OpenParen
	CloseParen
	Comma
	Equal
	Into
	OpenBrace
	CloseBrace
	Semi
	Newline
	Import
	Def
	Class
	Comment
	IVar
	Has
	Is
	If
	Inc
	Dec
	While
	UpDot
)

type Value struct {
	Type  Type
	Value interface{}
}

func isDigit(r rune) bool {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	default:
		return false
	}
}

func isTerm(err error) bool {
	switch err {
	case io.EOF:
		return true
	default:
		return false
	}
}

func isAtomRune(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}

	return false
}

func NewLexer(in string) (*Lexer, error) {
	lex := &Lexer{
		Source: in,
		r:      strings.NewReader(in),
	}
	return lex, nil
}

var ErrUnknownRune = errors.New("unknown rune")

func (l *Lexer) scanComment() (*Value, error) {
	var buf bytes.Buffer

	for {
		r, _, err := l.r.ReadRune()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		if r == '\n' {
			break
		}

		buf.WriteRune(r)
	}

	return &Value{Type: Comment, Value: buf.String()}, nil
}

func (l *Lexer) scanNewWord() (string, error) {
	r, _, err := l.r.ReadRune()
	if err != nil {
		return "", err
	}

	return l.scanWord(r)
}

func (l *Lexer) scanWord(r rune) (string, error) {
	var buf bytes.Buffer

	buf.WriteRune(r)

	for {
		r, _, err := l.r.ReadRune()
		if err != nil {
			break
		}

		cont := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'

		if !cont {
			l.r.UnreadRune()
			break
		}

		buf.WriteRune(r)
	}

	return buf.String(), nil
}

func (l *Lexer) scanOp(r rune) (*Value, error) {
	var buf bytes.Buffer

	buf.WriteRune(r)

	for {
		r, _, err := l.r.ReadRune()
		if err != nil {
			break
		}

		cont := unicode.IsPunct(r) || unicode.IsSymbol(r)

		if !cont {
			l.r.UnreadRune()
			break
		}

		buf.WriteRune(r)
	}

	return &Value{Type: Operator, Value: buf.String()}, nil
}

func (l *Lexer) scanDigit(width int) (int64, error) {
	var buf bytes.Buffer

	for i := 0; i < width; i++ {
		r, _, err := l.r.ReadRune()
		if err != nil {
			return 0, err
		}

		buf.WriteRune(r)
	}

	return strconv.ParseInt(buf.String(), 16, 64)
}

func (l *Lexer) scanExact(seq string) error {
	for _, x := range seq {
		r, _, err := l.r.ReadRune()
		if err != nil {
			return err
		}

		if x != r {
			return errors.Wrapf(ErrUnknownRune,
				"expected '%s' got '%s'", string(x), string(r))
		}
	}

	return nil
}

var Keywords = map[string]Type{
	"true":   True,
	"false":  False,
	"nil":    Nil,
	"import": Import,
	"def":    Def,
	"class":  Class,
	"has":    Has,
	"is":     Is,
	"if":     If,
	"while":  While,
}

func (l *Lexer) scanBare(r rune) (*Value, error) {
	word, err := l.scanWord(r)
	if err != nil {
		return nil, err
	}

	if t, ok := Keywords[word]; ok {
		return &Value{Type: t}, nil
	}

	return &Value{Type: Word, Value: word}, nil
}

func (l *Lexer) scanKeyword(r rune) (*Value, error) {
	switch r {
	case 'n':
		err := l.scanExact("il")
		if err != nil {
			return nil, err
		}

		return &Value{Type: Nil}, nil

	case 't':
		err := l.scanExact("rue")
		if err != nil {
			return nil, err
		}

		return &Value{Type: True}, nil
	case 'f':
		err := l.scanExact("alse")
		if err != nil {
			return nil, err
		}

		return &Value{Type: False}, nil
	default:
		return nil, errors.Wrapf(ErrUnknownRune, "rune: %s", string(r))
	}
}

func (l *Lexer) Next() (*Value, error) {
	var (
		r   rune
		err error
	)

	r = ' '

	for unicode.IsSpace(r) {
		r, _, err = l.r.ReadRune()
		if err != nil {
			if isTerm(err) {
				return &Value{Type: Term}, nil
			}

			return nil, err
		}

		if r == '\n' {
			return &Value{Type: Newline}, nil
		}
	}

	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		var buf bytes.Buffer

		base := 10

		if r == '0' {
			r, _, err := l.r.ReadRune()
			if err != nil {
				if isTerm(err) {
					return &Value{Type: Integer, Value: int64(0)}, nil
				}

				return nil, err
			}

			if r == 'x' {
				base = 16
			} else {
				buf.WriteRune('0')
				buf.WriteRune(r)
			}
		} else {
			buf.WriteRune(r)
		}

		for {
			r, _, err := l.r.ReadRune()
			if err == nil && isDigit(r) {
				buf.WriteRune(r)
				continue
			}

			if err != nil && !isTerm(err) {
				return nil, err
			}

			l.r.UnreadRune()

			i, err := strconv.ParseInt(buf.String(), base, 64)
			if err != nil {
				return nil, err
			}

			return &Value{
				Type:  Integer,
				Value: i,
			}, nil
		}
	case '"':
		var buf bytes.Buffer

		for {
			r, _, err := l.r.ReadRune()
			if err != nil {
				return nil, err
			}

			if r == '\\' {
				r, _, err := l.r.ReadRune()
				if err != nil {
					return nil, err
				}

				switch r {
				case 'n':
					buf.WriteByte('\n')
				case 'r':
					buf.WriteByte('\r')
				case 't':
					buf.WriteByte('\t')
				case 'u':
					i, err := l.scanDigit(4)
					if err != nil {
						return nil, err
					}

					buf.WriteRune(rune(i))
				case 'U':
					i, err := l.scanDigit(8)
					if err != nil {
						return nil, err
					}

					buf.WriteRune(rune(i))
				default:
					return nil, fmt.Errorf("Unknown escape: %s", string(r))
				}

				continue
			}

			if r != '"' {
				buf.WriteRune(r)
				continue
			}

			return &Value{
				Type:  String,
				Value: buf.String(),
			}, nil
		}
	case ':':
		var buf bytes.Buffer

		for {
			r, _, err := l.r.ReadRune()
			if err == nil && isAtomRune(r) {
				buf.WriteRune(r)
				continue
			}

			if err != nil && !isTerm(err) {
				return nil, err
			}

			l.r.UnreadRune()

			return &Value{Type: Atom, Value: buf.String()}, nil
		}
	case '.':
		r, _, err := l.r.ReadRune()
		if err != nil {
			return &Value{Type: Dot}, nil
		}

		if r == '^' {
			return &Value{Type: UpDot}, nil
		}

		l.r.UnreadRune()
		return &Value{Type: Dot}, nil
	case '(':
		return &Value{Type: OpenParen}, nil
	case ')':
		return &Value{Type: CloseParen}, nil
	case '{':
		return &Value{Type: OpenBrace}, nil
	case '}':
		return &Value{Type: CloseBrace}, nil
	case ',':
		return &Value{Type: Comma}, nil
	case ';':
		return &Value{Type: Semi}, nil
	case '\n':
		return &Value{Type: Newline}, nil
	case '@':
		word, err := l.scanNewWord()
		if err != nil {
			return nil, err
		}

		return &Value{Type: IVar, Value: word}, nil
	case '#':
		return l.scanComment()
	default:
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			op, err := l.scanOp(r)
			if err != nil {
				return op, err
			}

			switch op.Value {
			case "=":
				return &Value{Type: Equal}, nil
			case "=>":
				return &Value{Type: Into}, nil
			case "++":
				return &Value{Type: Inc}, nil
			case "--":
				return &Value{Type: Dec}, nil
			default:
				return op, nil
			}
		}

		return l.scanBare(r)
	}

	return &Value{}, nil
}
