package value

type Ref struct {
	Value Value
}

type Lambda struct {
	Code *Code
	Args int
	Refs []*Ref
}

func CreateLambda(code *Code, refs []*Ref, args int) *Lambda {
	return &Lambda{code, args, refs}
}

func (l *Lambda) Class(env Env) *Class {
	return nil
}
