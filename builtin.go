package main

import (
	"fmt"
	"strings"
)

type BuiltinFunc func(Env, []Expression) any

type Builtin struct {
	Arity int
	Func  BuiltinFunc
}

const ArityVariadic = -1

var builtins = map[string]Builtin{
	"println":     {ArityVariadic, builtinPrintln},
	"inputNumber": {1, builtinInputNumber},

	// "Arrays"
	"array": {0, builtinArray},
	"get":   {2, builtinGet},
	"push":  {2, builtinPush},
	"set":   {3, builtinSet},
	"len":   {1, builtinLen},
}

func builtinPrintln(env Env, e []Expression) any {
	var sb strings.Builder
	var v any
	for i, expr := range e {
		if i != 0 {
			sb.WriteString(", ")
		}
		v = expr(env)
		sb.WriteString(fmt.Sprintf("%v", v))
	}
	fmt.Println(sb.String())
	return v
}

func builtinInputNumber(env Env, e []Expression) any {
	return env["_inputNumbers"].([]int)[e[0](env).(int)]
}

func builtinArray(env Env, e []Expression) any {
	return &[]any{}
}

func builtinGet(env Env, e []Expression) any {
	// get(arr, 2)
	arr := e[0](env).(*[]any)
	idx := e[1](env).(int)
	return (*arr)[idx]
}

func builtinPush(env Env, e []Expression) any {
	// push(arr, 42)
	arr := e[0](env).(*[]any)
	v := e[1](env)
	*arr = append(*arr, v)
	return v
}

func builtinSet(env Env, e []Expression) any {
	// set(arr, 2, 42)
	arr := e[0](env).(*[]any)
	idx := e[1](env).(int)
	val := e[2](env)
	if idx < len(*arr) {
		(*arr)[idx] = val
	} else if idx == len(*arr) {
		*arr = append(*arr, val)
	} else {
		panic("index out of bounds")
	}
	return val
}

func builtinLen(env Env, e []Expression) any {
	// len(arr)
	return len(*(e[0](env).(*[]any)))
}
