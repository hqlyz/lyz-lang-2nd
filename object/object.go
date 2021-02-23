package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"lyz-lang/ast"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
	HASH_OBJ         = "HASH"
)

// Object interface
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Integer object
type Integer struct {
	Value int64
}

// Inspect function
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// Type function
func (i *Integer) Type() ObjectType { return ObjectType(INTEGER_OBJ) }

func (i *Integer) HashKey() HashKey { return HashKey{Type: i.Type(), Value: uint64(i.Value)} }

// Boolean object
type Boolean struct {
	Value bool
}

// Inspect function
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// Type function
func (b *Boolean) Type() ObjectType { return ObjectType(BOOLEAN_OBJ) }

func (b *Boolean) HashKey() HashKey {
	var v uint64
	if b.Value {
		v = 1
	} else {
		v = 0
	}
	return HashKey{Type: b.Type(), Value: v}
}

// Null object represents absence of value
type Null struct{}

// Inspect function
func (n *Null) Inspect() string { return "null" }

// Type function
func (n *Null) Type() ObjectType { return ObjectType(NULL_OBJ) }

// ReturnValue object
type ReturnValue struct {
	Value Object
}

// Inspect function
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

// Type function
func (rv *ReturnValue) Type() ObjectType { return ObjectType(RETURN_VALUE_OBJ) }

// Error object
type Error struct {
	Message string
}

// Inspect function
func (e *Error) Inspect() string { return "Error: " + e.Message }

// Type function
func (e *Error) Type() ObjectType { return ObjectType(ERROR_OBJ) }

// Environment object
type Environment struct {
	store map[string]Object
	outer *Environment
}

// NewEnclosedEnvironment function
func NewEnclosedEnvironment(outer *Environment) *Environment {
	newEnv := NewEnvironment()
	newEnv.outer = outer
	return newEnv
}

// NewEnvironment function
func NewEnvironment() *Environment {
	return &Environment{store: map[string]Object{}}
}

// Get function
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set function
func (e *Environment) Set(name string, obj Object) Object {
	e.store[name] = obj
	return obj
}

// Function object
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

// Type function
func (f *Function) Type() ObjectType { return FUNCTION_OBJ }

// Inspect function
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

// String object
type String struct {
	Value string
}

// Type function
func (s *String) Type() ObjectType { return STRING_OBJ }

// Inspect function
func (s *String) Inspect() string { return s.Value }

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// BuiltinFunction object
type BuiltinFunction func(args ...Object) Object

// Builtin object
type Builtin struct {
	Fn BuiltinFunction
}

// Type function
func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }

// Inspect function
func (b *Builtin) Inspect() string { return "builtin function" }

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out bytes.Buffer
	elems := []string{}
	for _, el := range a.Elements {
		elems = append(elems, el.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elems, ", "))
	out.WriteString("]")
	return out.String()
}

type Hashable interface {
	HashKey() HashKey
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

type HashPair struct {
	Key   Object
	Value Object
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, v := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", v.Key.Inspect(), v.Value.Inspect()))
	}

	out.WriteString("{")
	out.Write([]byte(strings.Join(pairs, ", ")))
	out.WriteString("}")

	return out.String()
}
