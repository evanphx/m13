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
	Name string
}

type Call struct {
	Receiver   Node
	MethodName string
	Args       []Node
}

type Assign struct {
	Name  string
	Value Node
}

type Lambda struct {
	Args []string
	Expr Node
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
