package value

func NewList(env Env, cap int) *List {
	list := &List{}
	list.SetClass(env.ListClass())

	list.data = make([]Value, 0, cap)

	return list
}

func (list *List) Append(v Value) {
	list.data = append(list.data, v)
}
