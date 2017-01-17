package ast

import "reflect"

func Rewrite(n Node, re func(Node) Node) {
	v := reflect.ValueOf(n).Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		sv := f.Interface()

		switch st := sv.(type) {
		case Node:
			nv := re(st)
			f.Set(reflect.ValueOf(nv))
			Rewrite(nv, re)
		case []Node:
			for j := 0; j < len(st); j++ {
				st[j] = re(st[j])
				Rewrite(st[j], re)
			}
		}
	}
}
