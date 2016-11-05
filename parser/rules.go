package parser

import (
	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/lex"
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

	expr := r.Rec()

	exprSep := r.Plus(r.Or(r.T(lex.Semi), r.T(lex.Newline)))

	exprAnother := r.F(r.Seq(exprSep, expr), r.Nth(1))

	exprList := r.Fs(
		r.Seq(expr, r.Star(exprAnother)),
		func(rv []RuleValue) RuleValue {
			if right, ok := rv[1].([]RuleValue); ok {
				return append([]RuleValue{rv[0]}, right...)
			} else {
				return rv[:1]
			}
		})

	prim := r.Or(
		r.Type(lex.Integer, func(lv *lex.Value) RuleValue {
			return &ast.Integer{lv.Value.(int64)}
		}),
		r.Type(lex.String, func(lv *lex.Value) RuleValue {
			return &ast.String{lv.Value.(string)}
		}),
		r.Type(lex.Atom, func(lv *lex.Value) RuleValue {
			return &ast.Atom{lv.Value.(string)}
		}),
		r.Type(lex.Word, func(lv *lex.Value) RuleValue {
			return &ast.Variable{lv.Value.(string)}
		}),
		r.Type(lex.IVar, func(lv *lex.Value) RuleValue {
			return &ast.IVar{lv.Value.(string)}
		}),
		r.Type(lex.True, func(lv *lex.Value) RuleValue {
			return &ast.True{}
		}),
		r.Type(lex.False, func(lv *lex.Value) RuleValue {
			return &ast.False{}
		}),
		r.Type(lex.Nil, func(lv *lex.Value) RuleValue {
			return &ast.Nil{}
		}),
	)

	attrAccess := r.Fs(
		r.Seq(expr, r.T(lex.Dot), r.T(lex.Word)),
		func(rv []RuleValue) RuleValue {
			return &ast.Attribute{
				Receiver: rv[0].(ast.Node),
				Name:     rv[2].(*lex.Value).Value.(string),
			}
		})

	primcall0 := r.Fs(
		r.Seq(expr, r.T(lex.Dot), r.T(lex.Word),
			r.T(lex.OpenParen), r.T(lex.CloseParen)),
		func(rv []RuleValue) RuleValue {
			return &ast.Call{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[2].(*lex.Value).Value.(string),
			}
		})

	anotherArg := r.F(r.Seq(r.T(lex.Comma), expr), r.Nth(1))

	argList := r.Fs(
		r.Seq(expr, r.Star(anotherArg)),
		func(rv []RuleValue) RuleValue {
			if right, ok := rv[1].([]RuleValue); ok {
				return append([]RuleValue{rv[0]}, right...)
			} else {
				return rv[:1]
			}
		})

	primcallN := r.Fs(
		r.Seq(expr, r.T(lex.Dot), r.T(lex.Word),
			r.T(lex.OpenParen), argList, r.T(lex.CloseParen)),
		func(rv []RuleValue) RuleValue {
			return &ast.Call{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[2].(*lex.Value).Value.(string),
				Args:       convert(rv[4].([]RuleValue)),
			}
		})

	braceBody := r.Fs(
		r.Seq(r.T(lex.OpenBrace), exprList, r.T(lex.CloseBrace)),
		func(rv []RuleValue) RuleValue {
			return &ast.Block{
				Expressions: convert(rv[1].([]RuleValue)),
			}
		})

	lambdaBody := r.Or(braceBody, expr)

	lambda0 := r.Fs(
		r.Seq(r.T(lex.Into), lambdaBody),
		func(rv []RuleValue) RuleValue {
			return &ast.Lambda{
				Expr: rv[1].(ast.Node),
			}
		})

	lambda1 := r.Fs(
		r.Seq(r.T(lex.Word), r.T(lex.Into), lambdaBody),
		func(rv []RuleValue) RuleValue {
			return &ast.Lambda{
				Expr: rv[2].(ast.Node),
				Args: []string{
					rv[0].(*lex.Value).Value.(string),
				},
			}
		})

	argDefListAnother := r.F(r.Seq(r.T(lex.Comma), r.T(lex.Word)), r.Nth(1))

	argDefListInner := r.Fs(
		r.Seq(r.T(lex.Word), r.Star(argDefListAnother)),
		func(rv []RuleValue) RuleValue {
			if right, ok := rv[1].([]RuleValue); ok {
				return append([]RuleValue{rv[0]}, right...)
			} else {
				return rv[:1]
			}
		})

	argDefList := r.Fs(
		r.Seq(r.T(lex.OpenParen), argDefListInner, r.T(lex.CloseParen)),
		func(rv []RuleValue) RuleValue {
			var args []string
			for _, arg := range rv[1].([]RuleValue) {
				args = append(args, arg.(*lex.Value).Value.(string))
			}

			return args
		})

	lambdaN := r.Fs(
		r.Seq(argDefList, r.T(lex.Into), lambdaBody),
		func(rv []RuleValue) RuleValue {
			return &ast.Lambda{
				Expr: rv[2].(ast.Node),
				Args: rv[0].([]string),
			}
		})

	expr.Rules = []Rule{
		lambdaN, lambda1, lambda0,
		primcallN, primcall0,
		attrAccess, prim,
	}

	stmt := r.Ref()

	stmtSep := r.Plus(r.Or(r.T(lex.Semi), r.T(lex.Newline)))

	stmtAnother := r.F(r.Seq(stmtSep, stmt), r.Nth(1))

	stmtList := r.Fs(
		r.Seq(r.Maybe(stmtSep), stmt, r.Star(stmtAnother)),
		func(rv []RuleValue) RuleValue {
			if right, ok := rv[2].([]RuleValue); ok {
				return append([]RuleValue{rv[1]}, right...)
			} else {
				return rv[1:2]
			}
		})

	attrAssign := r.Fs(
		r.Seq(expr, r.T(lex.Equal), expr),
		func(rv []RuleValue) RuleValue {
			switch sv := rv[0].(type) {
			case *ast.Variable:
				return &ast.Assign{
					Name:  sv.Name,
					Value: rv[2],
				}
			case *ast.Attribute:
				return &ast.AttributeAssign{
					Receiver: sv.Receiver,
					Name:     sv.Name,
					Value:    rv[2].(ast.Node),
				}
			default:
				panic("can't assign that")
			}
		})

	assign := r.Fs(
		r.Seq(r.T(lex.Word), r.T(lex.Equal), expr),
		func(rv []RuleValue) RuleValue {
			return &ast.Assign{
				Name:  rv[0].(*lex.Value).Value.(string),
				Value: rv[2],
			}
		})

	importRest := r.F(r.Seq(r.T(lex.Dot), r.T(lex.Word)), r.Nth(1))

	importPath := r.Fs(
		r.Seq(r.T(lex.Word), r.Star(importRest)),
		func(rv []RuleValue) RuleValue {
			var path []string

			path = append(path, rv[0].(*lex.Value).Value.(string))

			for _, part := range rv[1].([]RuleValue) {
				path = append(path, part.(*lex.Value).Value.(string))
			}

			return path
		})

	importR := r.Fs(
		r.Seq(r.T(lex.Import), importPath),
		func(rv []RuleValue) RuleValue {
			return &ast.Import{
				Path: rv[1].([]string),
			}
		})

	def := r.Fs(
		r.Seq(r.T(lex.Def), r.T(lex.Word), r.Maybe(argDefList), braceBody),
		func(rv []RuleValue) RuleValue {
			var args []string

			if x, ok := rv[2].([]string); ok {
				args = x
			}

			return &ast.Definition{
				Name:      rv[1].(*lex.Value).Value.(string),
				Arguments: args,
				Body:      rv[3].(ast.Node),
			}
		})

	classBody := r.Fs(
		r.Seq(r.T(lex.OpenBrace), stmtList, r.T(lex.CloseBrace)),
		func(rv []RuleValue) RuleValue {
			return &ast.Block{
				Expressions: convert(rv[1].([]RuleValue)),
			}
		})

	class := r.Fs(
		r.Seq(r.T(lex.Class), r.T(lex.Word), classBody),
		func(rv []RuleValue) RuleValue {
			return &ast.ClassDefinition{
				Name: rv[1].(*lex.Value).Value.(string),
				Body: rv[2].(ast.Node),
			}
		})

	comment := r.F(r.T(lex.Comment), func(rv RuleValue) RuleValue {
		return &ast.Comment{Comment: rv.(*lex.Value).Value.(string)}
	})

	is := r.Fs(
		r.Seq(r.T(lex.Is), r.T(lex.Word)),
		func(rv []RuleValue) RuleValue {
			return rv[1].(*lex.Value).Value.(string)
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
		r.Seq(r.T(lex.Has), r.T(lex.IVar), r.Maybe(hasTraits)),
		func(rv []RuleValue) RuleValue {
			var traits []string

			if x, ok := rv[2].([]string); ok {
				traits = x
			}

			return &ast.Has{
				Variable: rv[1].(*lex.Value).Value.(string),
				Traits:   traits,
			}
		})

	stmt.Rule = r.Or(comment, importR, class, def, has, attrAssign, assign, expr)

	p.root = r.Fs(
		r.Seq(stmtList, r.Maybe(stmtSep), r.T(lex.Term)),
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
