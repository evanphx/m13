package value

import (
	"context"
	"fmt"
)

type Cond interface {
	Match(env Env, recv Value, args []Value) bool
}

type CondFunc struct {
	Cond
	Method *Method
}

type CondDispatcher struct {
	Conditions []CondFunc
}

func (c *CondDispatcher) Exec(ctx context.Context, env Env, recv Value, args []Value) (Value, error) {
	for _, c := range c.Conditions {
		if c.Match(env, recv, args) {
			return c.Method.Func(ctx, env, recv, args)
		}
	}

	return nil, fmt.Errorf("No version found")
}

type CondAlways struct{}

func (_ CondAlways) Match(env Env, recv Value, args []Value) bool {
	return true
}

type CheckClass struct {
	cls *Class
	arg int
}

func NewCheckClass(cls *Class, arg int) *CheckClass {
	if cls == nil {
		panic("bad class")
	}

	return &CheckClass{cls, arg}
}

func (cc *CheckClass) Match(env Env, recv Value, args []Value) bool {
	if len(args) <= cc.arg {
		return false
	}

	val := args[cc.arg]

	if val.Class(env) == nil {
		fmt.Printf("%#v\n", val)
		panic("ug?")
	}

	// fmt.Printf("=> isa check, %T (%s) isa? %s => %v\n", val, val.Class(env).FullName(), cc.cls.FullName(), ISA(env, val, cc.cls))

	return ISA(env, val, cc.cls)
}
