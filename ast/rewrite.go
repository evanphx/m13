package ast

import "reflect"

func Rewrite(n Node, re func(Node) Node) Node {
	n = re(n)

	v := reflect.ValueOf(n).Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		sv := f.Interface()

		switch st := sv.(type) {
		case Node:
			f.Set(reflect.ValueOf(Rewrite(st, re)))
		case []Node:
			for j := 0; j < len(st); j++ {
				st[j] = Rewrite(st[j], re)
			}
		}
	}

	return n
}
