package builtin

import (
	"math/big"

	"github.com/evanphx/m13/value"
)

// m13
type Bool bool

// m13
type Integer struct{}

// m13 parent=Integer
type I64 int64

// m13 parent=Integer
type BigInt struct {
	I *big.Int
}

// m13 name=+
func (i I64) add(j I64) (I64, error) {
	return i + j, nil
}

// m13 name=++
func (i I64) inc() (I64, error) {
	return i + 1, nil
}

// m13 name=<
func (i I64) lt(j I64) (Bool, error) {
	if i < j {
		return true, nil
	}

	return false, nil
}

func MakeI64(i int64) value.Value {
	return I64(i)
}
