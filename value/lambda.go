package value

type Ref struct {
	Value Value
}

type Lambda struct {
	Object

	Code *Code
	Args int
	Self Value
	Refs []*Ref
}

func CreateLambda(env Env, code *Code, self Value, refs []*Ref, args int) *Lambda {
	lambda := &Lambda{
		Code: code,
		Args: args,
		Self: self,
		Refs: refs,
	}

	lambda.SetClass(env.LambdaClass())

	return lambda
}

func (l *Lambda) RedirectSelf(recv Value) *Lambda {
	var dup Lambda = *l

	dup.Self = recv

	return &dup
}
