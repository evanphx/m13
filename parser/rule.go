package parser

import "github.com/evanphx/m13/lex"

type Lexer interface {
	Next() *lex.Value
	Mark() int
	Rewind(int)
}

type Rule interface {
	Match(Lexer) (RuleValue, bool)
}

type RuleValue interface{}

type CodeRule struct {
	Rule Rule
	F    func(RuleValue) RuleValue
}

func (cr *CodeRule) Match(n Lexer) (RuleValue, bool) {
	v, ok := cr.Rule.Match(n)
	if !ok {
		return v, ok
	}

	return cr.F(v), true
}

type LexRule struct {
	Type lex.Type
	F    func(*lex.Value) RuleValue
}

func (lr *LexRule) Match(n Lexer) (RuleValue, bool) {
	v := n.Next()
	if v == nil {
		return nil, false
	}

	if v.Type != lr.Type {
		return nil, false
	}

	if lr.F != nil {
		return lr.F(v), true
	}

	return v, true
}

type AndRule struct {
	Rules []Rule

	F func([]RuleValue) (RuleValue, bool)
}

func (cr *AndRule) Match(n Lexer) (RuleValue, bool) {
	var values []RuleValue

	for _, rule := range cr.Rules {
		val, ok := rule.Match(n)
		if !ok {
			return nil, false
		}

		values = append(values, val)
	}

	return values, true
}

type OrRule struct {
	Rules []Rule
}

func (or *OrRule) Match(n Lexer) (RuleValue, bool) {
	m := n.Mark()

	for _, rule := range or.Rules {
		val, ok := rule.Match(n)
		if ok {
			return val, true
		}

		n.Rewind(m)
	}

	return nil, false
}

type RepeatRule struct {
	Rule Rule

	Star bool
}

func (rr *RepeatRule) Match(n Lexer) (RuleValue, bool) {
	var values []RuleValue

	for {
		m := n.Mark()

		val, ok := rr.Rule.Match(n)
		if !ok {
			n.Rewind(m)
			break
		}

		values = append(values, val)
	}

	if len(values) == 0 && !rr.Star {
		return nil, false
	}

	return values, true
}

type RefRule struct {
	Rule
}

type Rules struct{}

func (r *Rules) Seq(rules ...Rule) Rule {
	return &AndRule{Rules: rules}
}

func (r *Rules) Or(rules ...Rule) Rule {
	return &OrRule{Rules: rules}
}

func (r *Rules) Type(t lex.Type, f func(*lex.Value) RuleValue) Rule {
	return &LexRule{t, f}
}

func (r *Rules) T(t lex.Type) Rule {
	return &LexRule{t, nil}
}

func (r *Rules) F(x Rule, f func(RuleValue) RuleValue) Rule {
	return &CodeRule{x, f}
}

func (r *Rules) Fs(x Rule, f func([]RuleValue) RuleValue) Rule {
	w := func(r RuleValue) RuleValue {
		return f(r.([]RuleValue))
	}
	return &CodeRule{x, w}
}

func (r *Rules) Star(x Rule) Rule {
	return &RepeatRule{Rule: x, Star: true}
}

func (r *Rules) Plus(x Rule) Rule {
	return &RepeatRule{Rule: x}
}

func (r *Rules) Nth(n int) func(RuleValue) RuleValue {
	return func(r RuleValue) RuleValue {
		rv := r.([]RuleValue)
		return rv[n]
	}
}

func (r *Rules) Ref() *RefRule {
	return &RefRule{}
}
