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
	"inputLength": {0, builtinInputLength},
	"inputNumber": {1, builtinInputNumber},

	// "Arrays" & "Maps"
	"array": {0, builtinArray},
	"map":   {0, builtinMap},
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

func builtinInputLength(env Env, e []Expression) (any, error) {
	getInputNumbers := (env["_getInputNumbers"]).(func() ([]int, error))
	inputNumbers, err := getInputNumbers()
	if err != nil {
		return nil, err
	}
	return len(inputNumbers), nil
}

func builtinInputNumber(env Env, e []Expression) (any, error) {
	index, err := e[0](env)
	if err != nil {
		return nil, err
	}
	if i, ok := index.(int); ok {
		getInputNumbers := (env["_getInputNumbers"]).(func() ([]int, error))
		inputNumbers, err := getInputNumbers()
		if err != nil {
			return nil, err
		}
		return inputNumbers[i], nil
	} else {
		return nil, fmt.Errorf("argument needs to be a number, but was '%v'", index)
	}
}

func builtinArray(env Env, e []Expression) (any, error) {
	return &[]any{}, nil
}

func builtinMap(env Env, e []Expression) (any, error) {
	return &map[string]any{}, nil
}

func getSliceOrMap(env Env, e []Expression) (*[]any, *map[string]any, error) {
	v, err := e[0](env)
	if err != nil {
		return nil, nil, err
	}

	array, ok := v.(*[]any)
	if ok {
		return array, nil, nil
	}

	map_, ok := v.(*map[string]any)
	if ok {
		return nil, map_, nil
	}

	return nil, nil, fmt.Errorf("first argument needs to be an array or map, but was '%v'", v)
}

func getArrayIndex(env Env, e Expression) (int, error) {
	v, err := e(env)
	if err != nil {
		return 0, err
	}

	if i, ok := v.(int); ok {
		return i, nil
	} else {
		return 0, fmt.Errorf("second argument needs to be a number, but was '%v'", v)
	}
}

func getMapKey(env Env, e Expression) (string, error) {
	v, err := e(env)
	if err != nil {
		return "", err
	}

	if s, ok := v.(string); ok {
		return s, nil
	} else {
		return "", fmt.Errorf("second argument needs to be a string, but was '%v'", v)
	}
}

func arrayOrMapOp(env Env, e []Expression,
	sliceOp func(*[]any, int, Env, []Expression) (any, error),
	mapOp func(*map[string]any, string, Env, []Expression) (any, error)) (any, error) {
	slice, map_, err := getSliceOrMap(env, e)
	if err != nil {
		return nil, err
	} else if slice != nil {
		idx, err := getArrayIndex(env, e[1])
		if err != nil {
			return nil, err
		}

		return sliceOp(slice, idx, env, e)
	} else {
		key, err := getMapKey(env, e[1])
		if err != nil {
			return nil, err
		}

		return mapOp(map_, key, env, e)
	}
}

func builtinGet(env Env, e []Expression) (any, error) {
	// get(arr, 2) or get(arr, "hello")
	return arrayOrMapOp(env, e,
		func(slice *[]any, idx int, env Env, e []Expression) (any, error) {
			// get(arr, 2)
			return (*slice)[idx], nil
		}, func(map_ *map[string]any, key string, env Env, e []Expression) (any, error) {
			// get(arr, "hello")
			return (*map_)[key], nil
		},
	)
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
	// set(arr, 2, 42) or set(map, "hello", 42)
	return arrayOrMapOp(env, e,
		func(slice *[]any, idx int, env Env, e []Expression) (any, error) {
			// set(arr, 2, 42)
			v, err := e[2](env)
			if err != nil {
				return nil, err
			}
			if idx == len(*slice) {
				*slice = append(*slice, v)
			} else if idx < len(*slice) {
				(*slice)[idx] = v
			} else {
				return nil, fmt.Errorf("index %d out of bounds (length %d)", idx, len(*slice))
			}
			return v, nil
		}, func(map_ *map[string]any, key string, env Env, e []Expression) (any, error) {
			// set(arr, "hello", 42)
			v, err := e[2](env)
			if err != nil {
				return nil, err
			}
			(*map_)[key] = v
			return v, nil
		},
	)
}

func builtinLen(env Env, e []Expression) (any, error) {
	// len(arr)
	slice, map_, err := getSliceOrMap(env, e)
	if err != nil {
		return nil, err
	} else if slice != nil {
		return len(*slice), nil
	} else {
		return len(*map_), nil
	}
}
