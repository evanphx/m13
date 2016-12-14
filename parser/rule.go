package parser

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const debugApply = false

type Lexer interface {
	Mark() int
	Rewind(int)
	RuneScanner() io.RuneScanner
}

type Rule interface {
	Match(Lexer) (RuleValue, bool)
}

type memoValue struct {
	val RuleValue
	ok  bool
	pos int
}

type memoRecorder struct {
	m map[int]memoValue
}

func (m *memoRecorder) retrieve(n Lexer) (memoValue, bool) {
	if v, ok := m.m[n.Mark()]; ok {
		return v, true
	}

	return memoValue{}, false
}

func (m *memoRecorder) save(p int, n Lexer, v RuleValue, ok bool) {
	if m.m == nil {
		m.m = make(map[int]memoValue)
	}

	m.m[p] = memoValue{v, ok, n.Mark()}
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
	name string
}

func (r *RefRule) Name() string {
	return r.name
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
	p     *Parser
	name  string
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
		rv, ok := r.p.Apply(rule, n)
		if ok {
			return rv, ok
		}

		n.Rewind(p)
	}

	return nil, false
}

func (r *RecursiveRule) Name() string {
	return r.name
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

	if t.Max == 1 {
		if len(values) == 1 {
			return values[0], true
		} else {
			return nil, true
		}
	}

	return values, true
}

func (t *TimesRules) Name() string {
	if t.Min == 0 {
		if t.Max == 1 {
			return descRule(t.Rule) + "?"
		}
	}

	return descRule(t.Rule) + fmt.Sprintf("{%d..%d}", t.Min, t.Max)
}

type ScanRule struct {
	name string
	f    func(rs io.RuneScanner) (RuleValue, bool)
	m    memoRecorder
}

func (s *ScanRule) Match(n Lexer) (RuleValue, bool) {
	pos := n.Mark()

	if mv, ok := s.m.retrieve(n); ok {
		n.Rewind(mv.pos)
		return mv.val, mv.ok
	}

	v, ok := s.f(n.RuneScanner())
	if ok {
		s.m.save(pos, n, v, true)
	}

	return v, ok
}

func (s *ScanRule) Name() string {
	return s.name
}

type LiteralRule struct {
	memoRecorder
	literal string
}

func (s *LiteralRule) Match(n Lexer) (RuleValue, bool) {
	pos := n.Mark()

	if mv, ok := s.retrieve(n); ok {
		n.Rewind(mv.pos)
		return mv.val, mv.ok
	}

	rs := n.RuneScanner()

	for _, need := range s.literal {
		r, _, err := rs.ReadRune()
		if err != nil || need != r {
			n.Rewind(pos)
			return nil, false
		}
	}

	s.save(pos, n, s.literal, true)

	return s.literal, true
}

func (s *LiteralRule) Name() string {
	return fmt.Sprintf("%#v", s.literal)
}

type RegexpRule struct {
	src string
	pat *regexp.Regexp

	m memoRecorder
}

type saveReader struct {
	sub io.RuneReader
	buf bytes.Buffer
}

func (sr *saveReader) ReadRune() (rune, int, error) {
	r, i, err := sr.sub.ReadRune()
	if err == nil {
		sr.buf.WriteRune(r)
	}

	return r, i, err
}

func (r *RegexpRule) Match(n Lexer) (RuleValue, bool) {
	pos := n.Mark()

	if mv, ok := r.m.retrieve(n); ok {
		n.Rewind(mv.pos)
		return mv.val, mv.ok
	}

	sr := &saveReader{sub: n.RuneScanner()}

	res := r.pat.FindReaderSubmatchIndex(sr)
	if res == nil {
		n.Rewind(pos)
		return nil, false
	}

	if len(res) == 4 {
		res[0] = res[2]
		res[1] = res[3]
	}

	cursor := pos + res[1]
	capture := string(sr.buf.Bytes()[res[0]:res[1]])

	n.Rewind(cursor)

	r.m.save(pos, n, capture, true)

	return capture, true
}

func (r *RegexpRule) Name() string {
	return "/" + r.src + "/"
}

type NotRule struct {
	p *Parser
	r Rule
}

func (r *NotRule) Match(n Lexer) (RuleValue, bool) {
	defer n.Rewind(n.Mark())

	_, ok := r.p.Apply(r.r, n)
	return nil, !ok
}

func (r *NotRule) Name() string {
	return "!" + descRule(r.r)
}

type NoneRule struct{}

func (r *NoneRule) Match(n Lexer) (RuleValue, bool) {
	rs := n.RuneScanner()

	_, _, err := rs.ReadRune()
	if err == nil {
		rs.UnreadRune()
		return nil, false
	}

	return nil, true
}

func (r *NoneRule) Name() string {
	return "$"
}

type CheckRule struct {
	p *Parser
	r Rule
	f func(RuleValue) (RuleValue, bool)
}

func (r *CheckRule) Match(n Lexer) (RuleValue, bool) {
	v, ok := r.p.Apply(r.r, n)
	if ok {
		return r.f(v)
	}

	return nil, false
}

func (r *CheckRule) Name() string {
	return "check(" + descRule(r.r) + ")"
}

func (p *Parser) Apply(r Rule, n Lexer) (RuleValue, bool) {
	if !debugApply {
		return r.Match(n)
	}

	fmt.Printf("%02d @ %d => %s\n", p.applyDepth, n.Mark(), descRule(r))
	p.applyDepth++
	rv, ok := r.Match(n)
	p.applyDepth--
	if ok {
		fmt.Printf("%02d * %d => %s => %#v\n", p.applyDepth, n.Mark(), descRule(r), rv)
	}

	return rv, ok
}

type Rules struct {
	Parser *Parser
}

func (r *Rules) Rec(name string) *RecursiveRule {
	return &RecursiveRule{
		p:    r.Parser,
		name: name,
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

func (r *Rules) Ref(name string) *RefRule {
	return &RefRule{name: name}
}

func (r *Rules) Scan(name string, f func(io.RuneScanner) (RuleValue, bool)) *ScanRule {
	return &ScanRule{name: name, f: f}
}

func (r *Rules) S(lit string) *LiteralRule {
	return &LiteralRule{literal: lit}
}

func (r *Rules) Re(pat string) *RegexpRule {
	return &RegexpRule{src: pat, pat: regexp.MustCompile(`\A` + pat)}
}

func (r *Rules) Not(x Rule) *NotRule {
	return &NotRule{r.Parser, x}
}

func (r *Rules) None() *NoneRule {
	return &NoneRule{}
}

func (r *Rules) Check(x Rule, f func(v RuleValue) (RuleValue, bool)) Rule {
	return &CheckRule{r.Parser, x, f}
}

func descRule(r Rule) string {
	if nr, ok := r.(Named); ok {
		return nr.Name()
	}

	return fmt.Sprintf("%#v", r)
}
