package value

type Object struct {
	class *Class
}

func (o *Object) Class(env Env) *Class {
	return o.class
}
