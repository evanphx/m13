package value

import (
	"context"
	"hash/fnv"
	"sync"
)

type StringLiterals struct {
	lock    sync.Mutex
	strings map[string]*String
}

func (sl *StringLiterals) Lookup(env Env, s string) *String {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	if sl.strings == nil {
		sl.strings = make(map[string]*String)
	}

	if str, ok := sl.strings[s]; ok {
		return str
	}

	str := &String{String: s}
	str.SetClass(env.StringClass())

	sl.strings[s] = str

	return str
}

// m13
type String struct {
	Object
	String string
}

func (s *String) Hash() uint64 {
	h := fnv.New64()
	h.Write([]byte(s.String))
	return h.Sum64()
}

// m13 name=+
func (s *String) add(o *String) (*String, error) {
	var ret String = *s

	ret.String += o.String

	return &ret, nil
}

func initString(r *Package, cls *Class) {
	cls.AddMethod(&MethodDescriptor{
		Name: "==",
		Signature: Signature{
			Required: 1,
		},
		Func: func(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
			s1 := recv.(*String)
			s2, ok := args[0].(*String)
			if !ok {
				return env.False(), nil
			}

			if s1 == s2 {
				return env.True(), nil
			}

			if s1.String == s2.String {
				return env.True(), nil
			}

			return env.False(), nil
		},
	})
}
