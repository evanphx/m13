package parser

import (
	"bytes"
	"io"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/evanphx/m13/ast"
)

func convert(rv []RuleValue) []ast.Node {
	var nodes []ast.Node

	for _, r := range rv {
		nodes = append(nodes, r.(ast.Node))
	}

	return nodes
}

func (p *Parser) SetupRules() {
	var r Rules

	r.Parser = p

	expr := r.Rec("expr")

	isDigit := func(r rune) bool {
		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return true
		default:
			return false
		}
	}

	skip := r.Re(`[\s\t]*`)
	ws := r.Re(`[\s\t]+`)

	sym := func(s string) Rule {
		return r.Seq(r.S(s), skip)
	}

	kw := func(s string) Rule {
		return r.Seq(r.S(s), r.Not(r.Re("[a-zA-Z0-9_]")))
	}

	integer := r.Scan("integer", func(rs io.RuneScanner) (RuleValue, bool) {
		r, _, err := rs.ReadRune()

		if err != nil || !isDigit(r) {
			return nil, false
		}

		var buf bytes.Buffer

		base := 10

		if r == '0' {
			r, _, err := rs.ReadRune()
			if err != nil {
				return &ast.Integer{Value: int64(0)}, true
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
			r, _, err := rs.ReadRune()
			if err == nil && isDigit(r) {
				buf.WriteRune(r)
				continue
			}

			rs.UnreadRune()

			i, err := strconv.ParseInt(buf.String(), base, 64)
			if err != nil {
				return nil, false
			}

			return &ast.Integer{
				Value: i,
			}, true
		}
	})

	scanDigit := func(rs io.RuneScanner, width int) (int64, error) {
		var buf bytes.Buffer

		for i := 0; i < width; i++ {
			r, _, err := rs.ReadRune()
			if err != nil {
				return 0, err
			}

			buf.WriteRune(r)
		}

		return strconv.ParseInt(buf.String(), 16, 64)
	}

	qstring := r.Scan("qstring", func(rs io.RuneScanner) (RuleValue, bool) {
		r, _, err := rs.ReadRune()
		if err != nil || r != '"' {
			return nil, false
		}

		var buf bytes.Buffer

		for {
			r, _, err := rs.ReadRune()
			if err != nil {
				return nil, false
			}

			if r == '\\' {
				r, _, err := rs.ReadRune()
				if err != nil {
					return nil, false
				}

				switch r {
				case 'n':
					buf.WriteByte('\n')
				case 'r':
					buf.WriteByte('\r')
				case 't':
					buf.WriteByte('\t')
				case 'u':
					i, err := scanDigit(rs, 4)
					if err != nil {
						return nil, false
					}

					buf.WriteRune(rune(i))
				case 'U':
					i, err := scanDigit(rs, 8)
					if err != nil {
						return nil, false
					}

					buf.WriteRune(rune(i))
				default:
					// TODO expose the bad escape code?
					return nil, false
				}

				continue
			}

			if r != '"' {
				buf.WriteRune(r)
				continue
			}

			return &ast.String{
				Value: buf.String(),
			}, true
		}
	})

	atom := r.Scan("atom", func(rs io.RuneScanner) (RuleValue, bool) {
		r, _, err := rs.ReadRune()
		if err != nil || r != ':' {
			return nil, false
		}

		var buf bytes.Buffer

		for {
			r, _, err := rs.ReadRune()
			if err == nil && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
				buf.WriteRune(r)
				continue
			}

			if buf.Len() == 0 {
				return nil, false
			}

			return &ast.Atom{Value: buf.String()}, true
		}
	})

	rawword := r.Scan("word", func(rs io.RuneScanner) (RuleValue, bool) {
		var buf bytes.Buffer

		r, _, err := rs.ReadRune()
		if err != nil {
			return nil, false
		}

		if !(unicode.IsLetter(r) || r == '_') {
			rs.UnreadRune()
			return nil, false
		}

		buf.WriteRune(r)

		for {
			r, _, err := rs.ReadRune()
			if err != nil {
				break
			}

			cont := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'

			if !cont {
				rs.UnreadRune()
				break
			}

			buf.WriteRune(r)
		}

		return buf.String(), true
	})

	keywords := map[string]bool{
		"true":   true,
		"false":  true,
		"self":   true,
		"if":     true,
		"class":  true,
		"def":    true,
		"import": true,
		"nil":    true,
	}

	word := r.Check(rawword, func(v RuleValue) (RuleValue, bool) {
		word := v.(string)

		if _, isKw := keywords[word]; isKw {
			return nil, false
		}

		return word, true
	})

	opChars := [127]bool{}

	opChars['*'] = true
	opChars['+'] = true
	opChars['-'] = true
	opChars['='] = true

	opName := r.Scan("opName", func(rs io.RuneScanner) (RuleValue, bool) {
		r, _, err := rs.ReadRune()
		if err != nil {
			return nil, false
		}

		if r >= 127 || !opChars[r] {
			return nil, false
		}

		var buf bytes.Buffer

		buf.WriteRune(r)

		for {
			r, _, err := rs.ReadRune()
			if err != nil {
				break
			}

			cont := r < 127 && opChars[r]

			if !cont {
				rs.UnreadRune()
				break
			}

			buf.WriteRune(r)
		}

		switch buf.String() {
		case "=", "=>":
			return nil, false
		}

		return buf.String(), true
	})

	ivar := r.Re("@([a-zA-Z][a-zA-Z0-9_]*)")

	prim := r.Or(
		integer,
		qstring,
		atom,
		r.F(word, func(v RuleValue) RuleValue {
			return &ast.Variable{Name: v.(string)}
		}),
		r.F(ivar, func(v RuleValue) RuleValue {
			return &ast.IVar{v.(string)}
		}),
		r.F(r.S("true"), func(v RuleValue) RuleValue {
			return &ast.True{}
		}),
		r.F(r.S("false"), func(v RuleValue) RuleValue {
			return &ast.False{}
		}),
		r.F(r.S("nil"), func(v RuleValue) RuleValue {
			return &ast.Nil{}
		}),
		r.F(r.S("self"), func(v RuleValue) RuleValue {
			return &ast.Self{}
		}),
	)

	methodName := r.Re("[a-zA-Z_][a-zA-Z0-9_]*")

	dmc := r.F(r.Seq(r.S("."), methodName), r.Nth(1))
	dmuc := r.F(r.Seq(r.S(".^"), methodName), r.Nth(1))

	attrAccess := r.Fs(
		r.Seq(expr, dmc),
		func(rv []RuleValue) RuleValue {
			return &ast.Attribute{
				Receiver: rv[0].(ast.Node),
				Name:     rv[1].(string),
			}
		})

	primcall0 := r.Fs(
		r.Seq(expr, dmc, sym("("), sym(")")),
		func(rv []RuleValue) RuleValue {
			return &ast.Call{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[1].(string),
			}
		})

	anotherArg := r.F(r.Seq(r.S(","), skip, expr), r.Nth(2))

	argList := r.Fs(
		r.Seq(expr, skip, r.Star(anotherArg)),
		func(rv []RuleValue) RuleValue {
			if right, ok := rv[2].([]RuleValue); ok {
				return append([]RuleValue{rv[0]}, right...)
			} else {
				return rv[:1]
			}
		})

	primcallN := r.Fs(
		r.Seq(expr, dmc,
			sym("("), argList, sym(")")),
		func(rv []RuleValue) RuleValue {
			return &ast.Call{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[1].(string),
				Args:       convert(rv[3].([]RuleValue)),
			}
		})

	npcallN := r.Fs(
		r.Seq(expr, dmc, r.Re(`[\s\t]+`), argList),
		func(rv []RuleValue) RuleValue {
			return &ast.Call{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[1].(string),
				Args:       convert(rv[3].([]RuleValue)),
			}
		})
	upAttrAccess := r.Fs(
		r.Seq(expr, dmuc),
		func(rv []RuleValue) RuleValue {
			return &ast.UpCall{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[1].(string),
			}
		})

	upcall0 := r.Fs(
		r.Seq(expr, dmuc,
			sym("("), sym(")")),
		func(rv []RuleValue) RuleValue {
			return &ast.UpCall{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[1].(string),
			}
		})

	upcallN := r.Fs(
		r.Seq(expr, dmuc,
			sym("("), argList, sym(")")),
		func(rv []RuleValue) RuleValue {
			return &ast.UpCall{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[1].(string),
				Args:       convert(rv[3].([]RuleValue)),
			}
		})

	invoke := r.Fs(
		r.Seq(word, sym("("), argList, sym(")")),
		func(rv []RuleValue) RuleValue {
			return &ast.Invoke{
				Name: rv[0].(string),
				Args: convert(rv[2].([]RuleValue)),
			}
		})

	stmtList := r.Ref("stmtList")

	braceBody := r.Fs(
		r.Seq(sym("{"), stmtList, skip, sym("}")),
		func(rv []RuleValue) RuleValue {
			return &ast.Block{
				Expressions: convert(rv[1].([]RuleValue)),
			}
		})

	lambdaBody := r.Or(braceBody, expr)

	lambda0 := r.Fs(
		r.Seq(sym("=>"), lambdaBody),
		func(rv []RuleValue) RuleValue {
			return &ast.Lambda{
				Expr: rv[1].(ast.Node),
			}
		})

	lambda1 := r.Fs(
		r.Seq(word, skip, sym("=>"), lambdaBody),
		func(rv []RuleValue) RuleValue {
			return &ast.Lambda{
				Expr: rv[3].(ast.Node),
				Args: []string{
					rv[0].(string),
				},
			}
		})

	argDefListAnother := r.F(r.Seq(sym(","), word), r.Nth(1))

	argDefListInner := r.Fs(
		r.Seq(word, r.Star(argDefListAnother)),
		func(rv []RuleValue) RuleValue {
			if right, ok := rv[1].([]RuleValue); ok {
				return append([]RuleValue{rv[0]}, right...)
			} else {
				return rv[:1]
			}
		})

	argDefList := r.Fs(
		r.Seq(sym("("), argDefListInner, sym(")")),
		func(rv []RuleValue) RuleValue {
			var args []string
			for _, arg := range rv[1].([]RuleValue) {
				args = append(args, arg.(string))
			}

			return args
		})

	lambdaN := r.Fs(
		r.Seq(argDefList, sym("=>"), lambdaBody),
		func(rv []RuleValue) RuleValue {
			return &ast.Lambda{
				Expr: rv[2].(ast.Node),
				Args: rv[0].([]string),
			}
		})

	prec := map[string]int{
		"*":   4,
		"mul": 4,
		"/":   4,
		"div": 4,
		"+":   3,
		"add": 3,
		"-":   3,
		"sub": 3,
	}

	getPrec := func(op string) int {
		if v, ok := prec[op]; ok {
			return v
		}

		r, _ := utf8.DecodeRuneInString(op)

		if unicode.IsPunct(r) {
			if v, ok := prec[string(r)]; ok {
				return v
			}
		}

		return 0
	}

	op := r.Fs(
		r.Seq(expr, skip, opName, skip, expr),
		func(rv []RuleValue) RuleValue {
			op := rv[2].(string)

			if r, ok := rv[4].(*ast.Op); ok {
				if getPrec(op) > getPrec(r.Name) {
					return &ast.Op{
						Name: r.Name,
						Left: &ast.Op{
							Name:  op,
							Left:  rv[0].(ast.Node),
							Right: r.Left,
						},
						Right: r.Right,
					}
				}
			}

			return &ast.Op{
				Name:  op,
				Left:  rv[0].(ast.Node),
				Right: rv[4].(ast.Node),
			}
		})

	expr.Rules = []Rule{
		lambdaN, lambda1, lambda0,
		upcallN, upcall0, upAttrAccess,
		npcallN,
		op,
		primcallN, primcall0, invoke,
		attrAccess, prim,
	}

	stmt := r.Ref("stmt")

	stmtSep := r.Plus(r.Or(r.S(";"), r.S("\n")))

	stmtAnother := r.F(r.Seq(stmtSep, skip, stmt), r.Nth(2))

	stmtList.Rule = r.Fs(
		r.Seq(r.Maybe(stmtSep), skip, stmt, r.Star(stmtAnother), r.Maybe(stmtSep)),
		func(rv []RuleValue) RuleValue {
			if right, ok := rv[3].([]RuleValue); ok {
				return append([]RuleValue{rv[2]}, right...)
			} else {
				return rv[2:3]
			}
		})

	attrAssign := r.Fs(
		r.Seq(expr, skip, sym("="), skip, expr),
		func(rv []RuleValue) RuleValue {
			switch sv := rv[0].(type) {
			case *ast.Variable:
				return &ast.Assign{
					Name:  sv.Name,
					Value: rv[4].(ast.Node),
				}
			case *ast.Attribute:
				return &ast.AttributeAssign{
					Receiver: sv.Receiver,
					Name:     sv.Name,
					Value:    rv[4].(ast.Node),
				}
			default:
				panic("can't assign that")
			}
		})

	assign := r.Fs(
		r.Seq(word, skip, sym("="), skip, expr),
		func(rv []RuleValue) RuleValue {
			return &ast.Assign{
				Name:  rv[0].(string),
				Value: rv[4].(ast.Node),
			}
		})

	importRest := r.F(r.Seq(r.S("."), word), r.Nth(1))

	importPath := r.Fs(
		r.Seq(word, r.Star(importRest)),
		func(rv []RuleValue) RuleValue {
			var path []string

			path = append(path, rv[0].(string))

			for _, part := range rv[1].([]RuleValue) {
				path = append(path, part.(string))
			}

			return path
		})

	importR := r.Fs(
		r.Seq(kw("import"), ws, importPath),
		func(rv []RuleValue) RuleValue {
			return &ast.Import{
				Path: rv[2].([]string),
			}
		})

	def := r.Fs(
		r.Seq(kw("def"), ws, word, r.Maybe(argDefList), skip, braceBody),
		func(rv []RuleValue) RuleValue {
			var args []string

			if x, ok := rv[3].([]string); ok {
				args = x
			}

			return &ast.Definition{
				Name:      rv[2].(string),
				Arguments: args,
				Body:      rv[5].(ast.Node),
			}
		})

	classBody := r.Fs(
		r.Seq(sym("{"), stmtList, skip, sym("}")),
		func(rv []RuleValue) RuleValue {
			return &ast.Block{
				Expressions: convert(rv[1].([]RuleValue)),
			}
		})

	class := r.Fs(
		r.Seq(kw("class"), ws, word, skip, classBody),
		func(rv []RuleValue) RuleValue {
			return &ast.ClassDefinition{
				Name: rv[2].(string),
				Body: rv[4].(ast.Node),
			}
		})

	comment := r.F(r.Re(`#([^\n]*)`), func(rv RuleValue) RuleValue {
		return &ast.Comment{Comment: rv.(string)}
	})

	is := r.Fs(
		r.Seq(ws, kw("is"), ws, word),
		func(rv []RuleValue) RuleValue {
			return rv[3].(string)
		})

	hasTraits := r.F(
		r.Plus(is),
		func(rv RuleValue) RuleValue {
			var traits []string

			for _, r := range rv.([]RuleValue) {
				traits = append(traits, r.(string))
			}

			return traits
		})

	has := r.Fs(
		r.Seq(kw("has"), ws, ivar, r.Maybe(hasTraits)),
		func(rv []RuleValue) RuleValue {
			var traits []string

			if x, ok := rv[3].([]string); ok {
				traits = x
			}

			return &ast.Has{
				Variable: rv[2].(string),
				Traits:   traits,
			}
		})

	ifr := r.Fs(
		r.Seq(kw("if"), ws, expr, skip, braceBody),
		func(rv []RuleValue) RuleValue {
			return &ast.If{
				Cond: rv[2].(ast.Node),
				Body: rv[4].(ast.Node),
			}
		})

	while := r.Fs(
		r.Seq(kw("while"), ws, expr, skip, braceBody),
		func(rv []RuleValue) RuleValue {
			return &ast.While{
				Cond: rv[2].(ast.Node),
				Body: rv[4].(ast.Node),
			}
		})

	inc := r.Fs(
		r.Seq(expr, sym("++")),
		func(rv []RuleValue) RuleValue {
			return &ast.Inc{
				Receiver: rv[0].(ast.Node),
			}
		})

	dec := r.Fs(
		r.Seq(expr, sym("--")),
		func(rv []RuleValue) RuleValue {
			return &ast.Dec{
				Receiver: rv[0].(ast.Node),
			}
		})

	stmt.Rule = r.Or(comment, importR, class, def, has,
		ifr, while,
		attrAssign, assign, inc, dec,
		expr)

	p.expr = r.F(r.Seq(expr, r.None()), r.Nth(0))

	p.root = r.Fs(
		r.Seq(stmtList, r.Maybe(stmtSep), r.None()),
		func(rv []RuleValue) RuleValue {
			blk := rv[0].([]RuleValue)
			switch len(blk) {
			case 0:
				return nil
			case 1:
				return blk[0]
			default:
				return &ast.Block{Expressions: convert(blk)}
			}
		})
}
