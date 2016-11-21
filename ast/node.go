package ast

type Node interface{}

type Integer struct {
	Value int64
}

type String struct {
	Value string
}

type Atom struct {
	Value string
}

type True struct{}
type False struct{}
type Nil struct{}

type Variable struct {
	Name  string
	Ref   bool
	Index int
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

type Assign struct {
	Name  string
	Ref   bool
	Index int
	Value Node
}

type Lambda struct {
	Args  []string
	Scope *Scope
	Expr  Node
}

type Block struct {
	Expressions []Node
}

type Import struct {
	Path []string
}

type Attribute struct {
	Receiver Node
	Name     string
}

type AttributeAssign struct {
	Receiver Node
	Name     string
	Value    Node
}

type Definition struct {
	Name      string
	Arguments []string
	Body      Node
}

type ClassDefinition struct {
	Name string
	Body Node
}

type Comment struct {
	Comment string
}

type IVar struct {
	Name string
}

type Has struct {
	Variable string
	Traits   []string
}

type Op struct {
	Name  string
	Left  Node
	Right Node
}

type If struct {
	Cond Node
	Body Node
}

type Inc struct {
	Receiver Node
}

type Dec struct {
	Receiver Node
}

type While struct {
	Cond Node
	Body Node
}
