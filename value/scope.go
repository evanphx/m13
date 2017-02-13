package value

import "context"

type scopedVariable struct {
	name string
}

func GetScoped(ctx context.Context, name string) (Value, bool) {
	val := ctx.Value(scopedVariable{name})
	if val == nil {
		return nil, false
	}

	return val.(Value), true
}

func SetScoped(ctx context.Context, name string, val Value) context.Context {
	return context.WithValue(ctx, scopedVariable{name}, val)
}
