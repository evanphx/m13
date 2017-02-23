//go generate m13-glue

package value

import "math/big"

// m13
type Integer struct{}

func (i I64) Class(env Env) *Class {
	return env.I64Class()
}

func MakeI64(i int64) Value {
	return I64(i)
}

// m13 parent=Integer
type BigInt struct {
	I *big.Int
}
