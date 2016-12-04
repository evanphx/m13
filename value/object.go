package value

type Object struct {
	class *Class
}

func (o *Object) Class(env Env) *Class {
	return o.class
}

func (o *Object) SetClass(cls *Class) {
	o.class = cls
}
