package value

type Ref struct {
	Value Value
}

type Lambda struct {
	Object

	Code *Code
	Args int
	Refs []*Ref
}

func CreateLambda(env Env, code *Code, refs []*Ref, args int) *Lambda {
	lambda := &Lambda{
		Code: code,
		Args: args,
		Refs: refs,
	}

	// lambda.SetClass(env.LambdaClass())

	return lambda
}
