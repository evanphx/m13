package value

import (
	"context"
	"fmt"
)

type PackageRegistry struct {
	Packages map[string]*Package
}

type ClassConfig struct {
	Package    *Package
	Name       string
	GlobalName string
	Parent     string
	Methods    map[string]*Method
}

type Class struct {
	Object

	metaclass bool

	Parent     *Class
	Package    *Package
	Name       string
	GlobalName string
	Methods    map[string]*Method
	Ivars      map[string]int
}

func (c *Class) Class(env Env) *Class {
	return c.class
}

func (t *Class) FullName() string {
	return fmt.Sprintf("%s.%s", t.Package.Name, t.Name)
}

type MethodConfig struct {
	Name      string
	Signature Signature
	Func      func(ctx context.Context, env Env, recv Value, args []Value) (Value, error)
}

type Method struct {
	Name      string
	Signature Signature
	Object    interface{}
	Func      func(ctx context.Context, env Env, recv Value, args []Value) (Value, error)
}

type MethodDescriptor struct {
	Name      string
	Aliases   []string
	Signature Signature
	Object    interface{}
	Func      func(ctx context.Context, env Env, recv Value, args []Value) (Value, error)
}
