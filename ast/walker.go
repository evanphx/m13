package ast

import "reflect"

func Descend(n Node, re func(Node) bool) {
	if !re(n) {
		return
	}

	v := reflect.ValueOf(n).Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		sv := f.Interface()

		switch st := sv.(type) {
		case Node:
			Descend(st, re)
		case []Node:
			for j := 0; j < len(st); j++ {
				Descend(st[j], re)
			}
		}
	}
}
