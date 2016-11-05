package parser

import (
	"fmt"
	"strings"

	"github.com/evanphx/m13/lex"
)

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
	p    *Parser
	Rule Rule
	F    func(RuleValue) RuleValue
}

func (cr *CodeRule) Match(n Lexer) (RuleValue, bool) {
	v, ok := cr.p.Apply(cr.Rule, n)
	if !ok {
		return v, ok
	}

	return cr.F(v), true
}

func (cr *CodeRule) Name() string {
	if nr, ok := cr.Rule.(Named); ok {
		return nr.Name()
	}

	return descRule(cr.Rule)
}

type LexRule struct {
	p    *Parser
	Type lex.Type
	F    func(*lex.Value) RuleValue
}

func (lr *LexRule) Match(n Lexer) (RuleValue, bool) {
	v := n.Next()
	if v == nil {
		return nil, false
	}

	if v.Type != lr.Type {
		// lr.p.addAttempt(n.Mark(), lr.Type, v)
		return nil, false
	}

	if lr.F != nil {
		return lr.F(v), true
	}

	return v, true
}

func (lr *LexRule) GoString() string {
	return fmt.Sprintf("&parser.LexRule{Type:%s}", lr.Type)
}

func (lr *LexRule) Name() string {
	return fmt.Sprintf("match(%s)", lr.Type)
}

type AndRule struct {
	p *Parser

	Rules []Rule

	F func([]RuleValue) (RuleValue, bool)
}

func (cr *AndRule) Match(n Lexer) (RuleValue, bool) {
	var values []RuleValue

	for _, rule := range cr.Rules {
		val, ok := cr.p.Apply(rule, n)
		if !ok {
			return nil, false
		}

		values = append(values, val)
	}

	return values, true
}

func (cr *AndRule) Name() string {
	var parts []string

	for _, rule := range cr.Rules {
		if nr, ok := rule.(Named); ok {
			parts = append(parts, nr.Name())
		} else {
			parts = append(parts, fmt.Sprintf("%T(%p)", rule, rule))
		}
	}

	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

type OrRule struct {
	p     *Parser
	Rules []Rule
}

func (or *OrRule) Match(n Lexer) (RuleValue, bool) {
	m := n.Mark()

	for _, rule := range or.Rules {
		val, ok := or.p.Apply(rule, n)
		if ok {
			return val, true
		}

		n.Rewind(m)
	}

	return nil, false
}

func (or *OrRule) Name() string {
	var parts []string

	for _, rule := range or.Rules {
		if nr, ok := rule.(Named); ok {
			parts = append(parts, nr.Name())
		} else {
			parts = append(parts, fmt.Sprintf("%T(%p)", rule, rule))
		}
	}

	return fmt.Sprintf("(%s)", strings.Join(parts, " | "))
}

type RepeatRule struct {
	p    *Parser
	Rule Rule

	Star bool
}

func (rr *RepeatRule) Match(n Lexer) (RuleValue, bool) {
	var values []RuleValue

	for {
		m := n.Mark()

		val, ok := rr.p.Apply(rr.Rule, n)
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

func (rr *RepeatRule) Name() string {
	if rr.Star {
		return descRule(rr.Rule) + "*"
	}
	return descRule(rr.Rule) + "+"
}

type RefRule struct {
	Rule
}

type rrLR struct {
	detected bool
}

type rrMemo struct {
	pos int
	ans RuleValue
	ok  bool
}

type RecursiveRule struct {
	Rules []Rule
	memo  map[int]*rrMemo
}

func (rr *RecursiveRule) Add(r Rule) {
	rr.Rules = append(rr.Rules, r)
}

func (r *RecursiveRule) Match(n Lexer) (RuleValue, bool) {
	p := n.Mark()

	m, ok := r.memo[p]
	if !ok {
		lr := &rrLR{}
		m := &rrMemo{ans: lr, pos: p}
		r.memo[p] = m

		rv, ok := r.eval(n)
		m.ans = rv
		m.ok = ok
		m.pos = n.Mark()

		if lr.detected && ok {
			return r.growLR(n, p, m)
		}

		return rv, ok
	} else {
		n.Rewind(m.pos)
		if lr, ok := m.ans.(*rrLR); ok {
			lr.detected = true
			return nil, false
		}

		return m.ans, m.ok
	}
}

func (r *RecursiveRule) growLR(n Lexer, p int, m *rrMemo) (RuleValue, bool) {
	for {
		n.Rewind(p)

		ans, ok := r.eval(n)
		if !ok || n.Mark() <= m.pos {
			break
		}

		m.ans = ans
		m.ok = ok
		m.pos = n.Mark()
	}

	n.Rewind(m.pos)

	return m.ans, m.ok
}

func (r *RecursiveRule) eval(n Lexer) (RuleValue, bool) {
	p := n.Mark()

	for _, rule := range r.Rules {
		rv, ok := rule.Match(n)
		if ok {
			return rv, ok
		}

		n.Rewind(p)
	}

	return nil, false
}

type NamedRule struct {
	Rule
	name string
}

func (nr *NamedRule) Name() string {
	return nr.name
}

type Named interface {
	Name() string
}

type TimesRules struct {
	p    *Parser
	Rule Rule
	Min  int
	Max  int
}

func (t *TimesRules) Match(n Lexer) (RuleValue, bool) {
	var cnt int

	var values []RuleValue

	for cnt < t.Max {
		p := n.Mark()

		rv, ok := t.p.Apply(t.Rule, n)
		if !ok {
			n.Rewind(p)
			break
		}
		values = append(values, rv)
	}

	if cnt < t.Min {
		return nil, false
	}

	return values, true
}

func (p *Parser) Apply(r Rule, n Lexer) (RuleValue, bool) {
	return r.Match(n)
}

type Rules struct {
	Parser *Parser
}

func (r *Rules) Rec() *RecursiveRule {
	return &RecursiveRule{
		memo: make(map[int]*rrMemo),
	}
}

func (r *Rules) Name(name string, n Rule) *NamedRule {
	return &NamedRule{Rule: n, name: name}
}

func (r *Rules) Seq(rules ...Rule) Rule {
	return &AndRule{p: r.Parser, Rules: rules}
}

func (r *Rules) Or(rules ...Rule) Rule {
	return &OrRule{p: r.Parser, Rules: rules}
}

func (r *Rules) Type(t lex.Type, f func(*lex.Value) RuleValue) Rule {
	return &LexRule{r.Parser, t, f}
}

func (r *Rules) T(t lex.Type) Rule {
	return &LexRule{r.Parser, t, nil}
}

func (r *Rules) F(x Rule, f func(RuleValue) RuleValue) Rule {
	return &CodeRule{r.Parser, x, f}
}

func (r *Rules) Fs(x Rule, f func([]RuleValue) RuleValue) Rule {
	w := func(r RuleValue) RuleValue {
		return f(r.([]RuleValue))
	}
	return &CodeRule{r.Parser, x, w}
}

func (r *Rules) Star(x Rule) Rule {
	return &RepeatRule{p: r.Parser, Rule: x, Star: true}
}

func (r *Rules) Maybe(x Rule) Rule {
	return &TimesRules{p: r.Parser, Rule: x, Min: 0, Max: 1}
}

func (r *Rules) Plus(x Rule) Rule {
	return &RepeatRule{p: r.Parser, Rule: x}
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

func descRule(r Rule) string {
	if nr, ok := r.(Named); ok {
		return nr.Name()
	}

	return fmt.Sprintf("%#v", r)
}
