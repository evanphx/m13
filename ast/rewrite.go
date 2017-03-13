package ast

import "reflect"

func Rewrite(n Node, re func(Node) Node) Node {
	n = re(n)

	rv := reflect.ValueOf(n)

	v := rv.Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		sv := f.Interface()

		switch st := sv.(type) {
		case Node:
			if f.IsNil() {
				continue
			}

			f.Set(reflect.ValueOf(Rewrite(st, re)))
		case []Node:
			if f.IsNil() {
				continue
			}

			for j := 0; j < len(st); j++ {
				st[j] = Rewrite(st[j], re)
			}
		}
	}

	return n
}
