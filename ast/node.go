package ast

type Node interface {
	NodeType() string
}

type Integer struct {
	Value int64
}

func (i *Integer) NodeType() string {
	return "Integer"
}

type String struct {
	Value string
}

func (s *String) NodeType() string {
	return "string"
}

type Atom struct {
	Value string
}

func (a *Atom) NodeType() string {
	return "atom"
}

type True struct{}

func (v *True) NodeType() string {
	return "true"
}

type False struct{}

func (v *False) NodeType() string {
	return "false"
}

type Nil struct{}

func (v *Nil) NodeType() string {
	return "nil"
}

type Self struct{}

func (v *Self) NodeType() string {
	return "self"
}

type Variable struct {
	Name  string
	Ref   bool
	Index int
}

func (v *Variable) NodeType() string {
	return "variable"
}

type Scope struct {
	Locals []string
	Refs   []string
}

func (s *Scope) RefIndex(name string) int {
	for idx, ref := range s.Refs {
		if name == ref {
			return idx
		}
	}

	return -1
}

type Call struct {
	Receiver   Node
	MethodName string
	Args       []Node
}

func (v *Call) NodeType() string {
	return "call"
}

type UpCall struct {
	Receiver   Node
	MethodName string
	Args       []Node
}

func (v *UpCall) NodeType() string {
	return "upcall"
}

type Invoke struct {
	Var  Node
	Args []Node
}

func (v *Invoke) NodeType() string {
	return "invoke"
}

type Assign struct {
	Name  string
	Ref   bool
	Index int
	Value Node
}

func (v *Assign) NodeType() string {
	return "assign"
}

type Lambda struct {
	Name  string
	Args  []string
	Scope *Scope
	Expr  Node
}

func (v *Lambda) NodeType() string {
	return "lambda"
}

type Block struct {
	Expressions []Node
}

func (v *Block) NodeType() string {
	return "block"
}

type Import struct {
	Path []string
}

func (v *Import) NodeType() string {
	return "import"
}

type Attribute struct {
	Receiver Node
	Name     string
}

func (v *Attribute) NodeType() string {
	return "attribute"
}

type AttributeAssign struct {
	Receiver Node
	Name     string
	Value    Node
}

func (v *AttributeAssign) NodeType() string {
	return "attribute-assign"
}

type Definition struct {
	Name      string
	Arguments []string
	Body      Node
}

func (v *Definition) NodeType() string {
	return "def"
}

type ClassDefinition struct {
	Name string
	Body Node
}

func (v *ClassDefinition) NodeType() string {
	return "class"
}

type Comment struct {
	Comment string
}

func (v *Comment) NodeType() string {
	return "comment"
}

type ScopeVar struct {
	Name string
}

func (v *ScopeVar) NodeType() string {
	return "scopevar"
}

type IVar struct {
	Name string
}

func (v *IVar) NodeType() string {
	return "ivar"
}

type IVarAssign struct {
	Name  string
	Index int
	Value Node
}

func (v *IVarAssign) NodeType() string {
	return "ivarassign"
}

type Has struct {
	Variable string
	Traits   []string
}

func (v *Has) NodeType() string {
	return "has"
}

type Op struct {
	Name  string
	Left  Node
	Right Node
}

func (v *Op) NodeType() string {
	return "op"
}

type If struct {
	Cond Node
	Body Node
	Else Node
}

func (v *If) NodeType() string {
	return "if"
}

type Inc struct {
	Receiver Node
}

func (v *Inc) NodeType() string {
	return "inc"
}

type Dec struct {
	Receiver Node
}

func (v *Dec) NodeType() string {
	return "dec"
}

type While struct {
	Cond Node
	Body Node
}

func (v *While) NodeType() string {
	return "while"
}

type List struct {
	Elements []Node
}

func (l *List) NodeType() string {
	return "list"
}
