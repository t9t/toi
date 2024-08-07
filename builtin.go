package main

import (
	"fmt"
	"strings"
)

type BuiltinFunc func(Env, []Expression) (any, error)

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

func builtinPrintln(env Env, e []Expression) (any, error) {
	var sb strings.Builder
	for i, expr := range e {
		if i != 0 {
			sb.WriteString(", ")
		}
		v, err := expr(env)
		if err != nil {
			return nil, err
		}
		sb.WriteString(fmt.Sprintf("%v", v))
	}
	fmt.Println(sb.String())
	return nil, nil
}

func builtinInputNumber(env Env, e []Expression) (any, error) {
	index, err := e[0](env)
	if err != nil {
		return nil, err
	}
	if i, ok := index.(int); ok {
		return env["_inputNumbers"].([]int)[i], nil
	} else {
		return nil, fmt.Errorf("argument needs to be a number, but was '%v'", index)
	}
}

func builtinArray(env Env, e []Expression) (any, error) {
	return &[]any{}, nil
}

func builtinGet(env Env, e []Expression) (any, error) {
	// get(arr, 2)
	arr, err := e[0](env)
	if err != nil {
		return nil, err
	}

	array, ok := arr.(*[]any)
	if !ok {
		return nil, fmt.Errorf("first argument needs to be an array, but was '%v'", arr)
	}

	idx, err := e[1](env)
	if err != nil {
		return nil, err
	}

	if i, ok := idx.(int); ok {
		return (*array)[i], nil
	} else {
		return nil, fmt.Errorf("second argument needs to be a number, but was '%v'", idx)
	}
}

func builtinPush(env Env, e []Expression) (any, error) {
	// push(arr, 42)
	arr, err := e[0](env)
	if err != nil {
		return nil, err
	}

	array, ok := arr.(*[]any)
	if !ok {
		return nil, fmt.Errorf("first argument needs to be an array, but was '%v'", arr)
	}

	if v, err := e[1](env); err != nil {
		return nil, err
	} else {
		*array = append(*array, v)
		return v, nil
	}
}

func builtinSet(env Env, e []Expression) (any, error) {
	// set(arr, 2, 42)
	arr, err := e[0](env)
	if err != nil {
		return nil, err
	}

	array, ok := arr.(*[]any)
	if !ok {
		return nil, fmt.Errorf("first argument needs to be an array, but was '%v'", arr)
	}

	idx, err := e[1](env)
	if err != nil {
		return nil, err
	}

	i, ok := idx.(int)
	if !ok {
		return nil, fmt.Errorf("second argument needs to be a number, but was '%v'", idx)
	}

	v, err := e[2](env)
	if err != nil {
		return nil, err
	}

	if i < len(*array) {
		(*array)[i] = v
	} else if i == len(*array) {
		*array = append(*array, v)
	} else {
		return nil, fmt.Errorf("index %d out of bounds (length %d)", i, len(*array))
	}
	return v, nil
}

func builtinLen(env Env, e []Expression) (any, error) {
	// len(arr)
	arr, err := e[0](env)
	if err != nil {
		return nil, err
	}

	array, ok := arr.(*[]any)
	if !ok {
		return nil, fmt.Errorf("first argument needs to be an array, but was '%v'", arr)
	}

	return len(*array), nil
}
