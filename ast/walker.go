package ast

import "reflect"

func Descend(n Node, re func(Node) bool) {
	if !re(n) {
		return
	}

	rv := reflect.ValueOf(n)
	if rv.IsNil() {
		return
	}

	v := rv.Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		sv := f.Interface()

		switch st := sv.(type) {
		case Node:
			if f.IsNil() {
				continue
			}

			Descend(st, re)
		case []Node:
			if f.IsNil() {
				continue
			}

			for j := 0; j < len(st); j++ {
				Descend(st[j], re)
			}
		}
	}
}
