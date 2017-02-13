package vm

import (
	"fmt"

	"github.com/evanphx/m13/value"
)

type Cond interface {
	Match(env value.Env, recv value.Value, args []value.Value) bool
}

type CondFunc struct {
	Cond
	F func(env value.Env, recv value.Value, args []value.Value) (value.Value, error)
}

type CondDispatcher struct {
	Conditions []CondFunc
}

func (c *CondDispatcher) Exec(env value.Env, recv value.Value, args []value.Value) (value.Value, error) {
	for _, c := range c.Conditions {
		if c.Match(env, recv, args) {
			return c.F(env, recv, args)
		}
	}

	return nil, fmt.Errorf("No version found")
}

type CheckClass struct {
	cls *value.Class
	arg int
}

func (cc *CheckClass) Match(env value.Env, recv value.Value, args []value.Value) bool {
	if len(args) <= cc.arg {
		return false
	}

	val := args[cc.arg]

	return cc.cls == val.Class(env)
}
