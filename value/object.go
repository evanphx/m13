package value

import "unsafe"

type Object struct {
	class *Class
}

func (o *Object) Class(env Env) *Class {
	return o.class
}

func (o *Object) SetClass(cls *Class) {
	o.class = cls
}

func (o *Object) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(o)))
}
