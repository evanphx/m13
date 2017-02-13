package value

import "fmt"

func Inspect(env Env, val Value) string {
	if val == nil {
		return "<nil>"
	}

	switch sv := val.(type) {
	case *String:
		return fmt.Sprintf(`"%s"`, sv.String)
	default:
		return fmt.Sprintf("<%s:%p>", val.Class(env).GlobalName, val)
	}
}
