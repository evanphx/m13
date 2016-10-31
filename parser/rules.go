package parser

import (
	"github.com/evanphx/m13/ast"
	"github.com/evanphx/m13/lex"
)

func (p *Parser) SetupRules() {
	var r Rules

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

	expr := r.Ref()

	arg := expr

	primcallNA := r.Fs(
		r.Seq(prim, r.T(lex.Dot), r.T(lex.Word),
			r.T(lex.OpenParen), r.T(lex.CloseParen)),
		func(rv []RuleValue) RuleValue {
			return &ast.Call{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[2].(*lex.Value).Value.(string),
			}
		})

	anotherArg := r.F(r.Seq(r.T(lex.Comma), arg), r.Nth(1))

	argList := r.Fs(
		r.Seq(arg, r.Star(anotherArg)),
		func(rv []RuleValue) RuleValue {
			if right, ok := rv[1].([]RuleValue); ok {
				return append([]RuleValue{rv[0]}, right...)
			} else {
				return rv[:1]
			}
		})

	convert := func(rv []RuleValue) []ast.Node {
		var nodes []ast.Node

		for _, r := range rv {
			nodes = append(nodes, r.(ast.Node))
		}

		return nodes
	}

	primcallA := r.Fs(
		r.Seq(prim, r.T(lex.Dot), r.T(lex.Word),
			r.T(lex.OpenParen), argList, r.T(lex.CloseParen)),
		func(rv []RuleValue) RuleValue {
			return &ast.Call{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[2].(*lex.Value).Value.(string),
				Args:       convert(rv[4].([]RuleValue)),
			}
		})

	primcall := r.Or(primcallA, primcallNA)

	chainedNA := r.Fs(
		r.Seq(
			primcall, r.T(lex.Dot), r.T(lex.Word),
			r.T(lex.OpenParen), r.T(lex.CloseParen)),
		func(rv []RuleValue) RuleValue {
			return &ast.Call{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[2].(*lex.Value).Value.(string),
			}
		})

	chainedA := r.Fs(
		r.Seq(
			primcall, r.T(lex.Dot), r.T(lex.Word),
			r.T(lex.OpenParen), argList, r.T(lex.CloseParen)),
		func(rv []RuleValue) RuleValue {
			return &ast.Call{
				Receiver:   rv[0].(ast.Node),
				MethodName: rv[2].(*lex.Value).Value.(string),
				Args:       convert(rv[4].([]RuleValue)),
			}
		})

	chained := r.Or(chainedA, chainedNA)

	attrRecv := r.Or(primcall, prim)

	attrAccess := r.Fs(
		r.Seq(attrRecv, r.T(lex.Dot), r.T(lex.Word)),
		func(rv []RuleValue) RuleValue {
			return &ast.Attribute{
				Receiver: rv[0].(ast.Node),
				Name:     rv[2].(*lex.Value).Value.(string),
			}
		})

	attrAssign := r.Fs(
		r.Seq(attrRecv, r.T(lex.Dot), r.T(lex.Word), r.T(lex.Equal), expr),
		func(rv []RuleValue) RuleValue {
			return &ast.AttributeAssign{
				Receiver: rv[0].(ast.Node),
				Name:     rv[2].(*lex.Value).Value.(string),
				Value:    rv[4].(ast.Node),
			}
		})

	attr := r.Or(attrAssign, attrAccess)

	assign := r.Fs(
		r.Seq(r.T(lex.Word), r.T(lex.Equal), expr),
		func(rv []RuleValue) RuleValue {
			return &ast.Assign{
				Name:  rv[0].(*lex.Value).Value.(string),
				Value: rv[2],
			}
		})

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

	braceBody := r.Fs(
		r.Seq(r.T(lex.OpenBrace), exprList, r.T(lex.CloseBrace)),
		func(rv []RuleValue) RuleValue {
			return &ast.Block{
				Expressions: convert(rv[1].([]RuleValue)),
			}
		})

	lambdaBody := r.Or(braceBody, expr)

	lambdaNA := r.Fs(
		r.Seq(r.T(lex.Into), lambdaBody),
		func(rv []RuleValue) RuleValue {
			return &ast.Lambda{
				Expr: rv[1].(ast.Node),
			}
		})

	lambda1A := r.Fs(
		r.Seq(r.T(lex.Word), r.T(lex.Into), lambdaBody),
		func(rv []RuleValue) RuleValue {
			return &ast.Lambda{
				Expr: rv[2].(ast.Node),
				Args: []string{
					rv[0].(*lex.Value).Value.(string),
				},
			}
		})

	lambdaAnotherArg := r.F(r.Seq(r.T(lex.Comma), r.T(lex.Word)), r.Nth(1))

	lambdaArgList := r.Fs(
		r.Seq(r.T(lex.Word), r.Star(lambdaAnotherArg)),
		func(rv []RuleValue) RuleValue {
			if right, ok := rv[1].([]RuleValue); ok {
				return append([]RuleValue{rv[0]}, right...)
			} else {
				return rv[:1]
			}
		})

	lambdaMA := r.Fs(
		r.Seq(r.T(lex.OpenParen), lambdaArgList, r.T(lex.CloseParen),
			r.T(lex.Into), lambdaBody),
		func(rv []RuleValue) RuleValue {
			var args []string
			for _, arg := range rv[1].([]RuleValue) {
				args = append(args, arg.(*lex.Value).Value.(string))
			}

			return &ast.Lambda{
				Expr: rv[4].(ast.Node),
				Args: args,
			}
		})

	lambda := r.Or(lambdaMA, lambda1A, lambdaNA)

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

	expr.Rule = r.Or(importR, assign, lambda, chained, attr, primcall, prim)

	p.root = r.Fs(
		r.Seq(exprList, r.T(lex.Term)),
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
