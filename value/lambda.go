package value

type Lambda struct {
	Code *Code
	Args int
}

func CreateLambda(code *Code, args int) *Lambda {
	return &Lambda{code, args}
}

func (l *Lambda) Type(env Env) *Type {
	return nil
}
