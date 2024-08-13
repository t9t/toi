package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type BuiltinFunc func(Env, []Expression) (any, error)
type BuiltinVmFunc func([]any) (any, error)

type Builtin struct {
	Arity  int
	Func   BuiltinFunc
	VmFunc BuiltinVmFunc
}

const ArityVariadic = -1

var builtins = map[string]Builtin{
	"println":      {ArityVariadic, builtinPrintln, builtinPrintlnVm},
	"inputLength":  {0, builtinInputLength, builtinInputLengthVm},
	"inputNumbers": {0, builtinInputNumbers, builtinInputNumbersVm},
	"inputLines":   {0, builtinInputLines, builtinInputLinesVm},

	"split": {2, builtinSplit, builtinSplitVm},
	"chars": {1, builtinChars, builtinCharsVm},

	"int":    {1, builtinInt, builtinIntVm},
	"string": {1, builtinString, builtinStringVm},

	// "Arrays" & "Maps"
	"array": {0, builtinArray, builtinArrayVm},
	"map":   {0, builtinMap, builtinMapVm},
	"get":   {2, builtinGet, builtinGetVm},
	"push":  {2, builtinPush, builtinPushVm},
	"set":   {3, builtinSet, builtinSetVm},
	"len":   {1, builtinLen, builtinLenVm},
	"keys":  {1, builtinKeys, builtinKeysVm},
}

func builtinPrintln(env Env, e []Expression) (any, error) {
	for i, expr := range e {
		if i != 0 {
			toiStdout.WriteString(", ")
		}
		v, err := expr.evaluate(env)
		if err != nil {
			return nil, err
		}
		if array, ok := v.(*[]any); ok {
			toiStdout.WriteRune('[')
			for i, element := range *array {
				if i != 0 {
					toiStdout.WriteString(", ")
				}
				toiStdout.WriteString(fmt.Sprintf("%v", element))
			}
			toiStdout.WriteRune(']')
		} else {
			toiStdout.WriteString(fmt.Sprintf("%v", v))
		}
	}
	toiStdout.WriteRune('\n')
	return nil, nil
}

func builtinPrintlnVm(arguments []any) (any, error) {
	for i, v := range arguments {
		if i != 0 {
			toiStdout.WriteString(", ")
		}
		if array, ok := v.(*[]any); ok {
			toiStdout.WriteRune('[')
			for i, element := range *array {
				if i != 0 {
					toiStdout.WriteString(", ")
				}
				toiStdout.WriteString(fmt.Sprintf("%v", element))
			}
			toiStdout.WriteRune(']')
		} else {
			toiStdout.WriteString(fmt.Sprintf("%v", v))
		}
	}
	toiStdout.WriteRune('\n')
	return nil, nil
}

func builtinInputLength(env Env, e []Expression) (any, error) {
	inputNumbers, err := getInputNumbers()
	if err != nil {
		return nil, err
	}
	return len(inputNumbers), nil
}

func builtinInputLengthVm(arguments []any) (any, error) {
	inputNumbers, err := getInputNumbers()
	if err != nil {
		return nil, err
	}
	return len(inputNumbers), nil
}

func getInputNumbers() ([]int, error) {
	inputNumbers := make([]int, 0)
	for _, line := range strings.Split(toiStdin, "\n") {
		if line == "" {
			continue
		}
		n, err := strconv.Atoi(line)
		if err != nil {
			return nil, err
		}
		inputNumbers = append(inputNumbers, n)
	}
	return inputNumbers, nil
}

func builtinInputLines(env Env, e []Expression) (any, error) {
	return toToiArray(strings.Split(strings.TrimSpace(toiStdin), "\n")), nil
}

func builtinInputLinesVm(arguments []any) (any, error) {
	return toToiArray(strings.Split(strings.TrimSpace(toiStdin), "\n")), nil
}

func builtinInputNumbers(env Env, e []Expression) (any, error) {
	numbers, err := getInputNumbers()
	if err != nil {
		return nil, err
	}
	return toToiArray(numbers), nil
}

func builtinInputNumbersVm(arguments []any) (any, error) {
	numbers, err := getInputNumbers()
	if err != nil {
		return nil, err
	}
	return toToiArray(numbers), nil
}

func builtinSplit(env Env, e []Expression) (any, error) {
	var maybeStr, maybeSep any
	var str, sep string
	var ok bool
	var err error

	if maybeStr, err = e[0].evaluate(env); err != nil {
		return nil, err
	} else if maybeSep, err = e[1].evaluate(env); err != nil {
		return nil, err
	}

	if str, ok = maybeStr.(string); !ok {
		return nil, fmt.Errorf("first argument needs to be a string, but was '%v'", maybeStr)
	} else if sep, ok = maybeSep.(string); !ok {
		return nil, fmt.Errorf("second argument needs to be a string, but was '%v'", maybeSep)
	}

	return toToiArray(strings.Split(str, sep)), nil
}

func builtinSplitVm(arguments []any) (any, error) {
	maybeStr, maybeSep := arguments[0], arguments[1]
	var str, sep string
	var ok bool

	if str, ok = maybeStr.(string); !ok {
		return nil, fmt.Errorf("first argument needs to be a string, but was '%v'", maybeStr)
	} else if sep, ok = maybeSep.(string); !ok {
		return nil, fmt.Errorf("second argument needs to be a string, but was '%v'", maybeSep)
	}

	return toToiArray(strings.Split(str, sep)), nil
}

func builtinChars(env Env, e []Expression) (any, error) {
	v, err := e[0].evaluate(env)
	if err != nil {
		return nil, err
	}

	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("argument needs to be a string, but was '%v'", v)
	}

	return toToiArray(strings.Split(s, "")), nil
}

func builtinCharsVm(arguments []any) (any, error) {
	v := arguments[0]
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("argument needs to be a string, but was '%v'", v)
	}

	return toToiArray(strings.Split(s, "")), nil
}

func builtinString(env Env, e []Expression) (any, error) {
	v, err := e[0].evaluate(env)
	if err != nil {
		return nil, err
	}
	if i, ok := v.(int); !ok {
		return nil, fmt.Errorf("argument needs to be an int, but was '%v'", v)
	} else {
		return strconv.Itoa(i), nil
	}
}

func builtinStringVm(arguments []any) (any, error) {
	v := arguments[0]
	if i, ok := v.(int); !ok {
		return nil, fmt.Errorf("argument needs to be an int, but was '%v'", v)
	} else {
		return strconv.Itoa(i), nil
	}
}

func builtinInt(env Env, e []Expression) (any, error) {
	v, err := e[0].evaluate(env)
	if err != nil {
		return nil, err
	}
	if s, ok := v.(string); !ok {
		return nil, fmt.Errorf("argument needs to be a string, but was '%v'", v)
	} else {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		return i, nil
	}
}

func builtinIntVm(arguments []any) (any, error) {
	v := arguments[0]
	if s, ok := v.(string); !ok {
		return nil, fmt.Errorf("argument needs to be a string, but was '%v'", v)
	} else {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		return i, nil
	}
}

func builtinArray(env Env, e []Expression) (any, error) {
	return &[]any{}, nil
}

func builtinArrayVm(arguments []any) (any, error) {
	return &[]any{}, nil
}

func builtinMap(env Env, e []Expression) (any, error) {
	return &map[string]any{}, nil
}

func builtinMapVm(arguments []any) (any, error) {
	return &map[string]any{}, nil
}

func getSliceOrMap(env Env, e []Expression) (*[]any, *map[string]any, error) {
	v, err := e[0].evaluate(env)
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

func getSliceOrMapVm(arguments []any) (*[]any, *map[string]any, error) {
	v := arguments[0]

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
	v, err := e.evaluate(env)
	if err != nil {
		return 0, err
	}

	if i, ok := v.(int); ok {
		return i, nil
	} else {
		return 0, fmt.Errorf("second argument needs to be a number, but was '%v'", v)
	}
}

func getArrayIndexVm(v any) (int, error) {
	if i, ok := v.(int); ok {
		return i, nil
	} else {
		return 0, fmt.Errorf("second argument needs to be a number, but was '%v'", v)
	}
}

func getMapKey(env Env, e Expression) (string, error) {
	v, err := e.evaluate(env)
	if err != nil {
		return "", err
	}

	if s, ok := v.(string); ok {
		return s, nil
	} else {
		return "", fmt.Errorf("second argument needs to be a string, but was '%v'", v)
	}
}

func getMapKeyVm(v any) (string, error) {
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

func arrayOrMapOpVm(arguments []any,
	sliceOp func(*[]any, int, []any) (any, error),
	mapOp func(*map[string]any, string, []any) (any, error)) (any, error) {
	slice, map_, err := getSliceOrMapVm(arguments)
	if err != nil {
		return nil, err
	} else if slice != nil {
		idx, err := getArrayIndexVm(arguments[1])
		if err != nil {
			return nil, err
		}

		return sliceOp(slice, idx, arguments)
	} else {
		key, err := getMapKeyVm(arguments[1])
		if err != nil {
			return nil, err
		}

		return mapOp(map_, key, arguments)
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

func builtinGetVm(arguments []any) (any, error) {
	// get(arr, 2) or get(arr, "hello")
	return arrayOrMapOpVm(arguments,
		func(slice *[]any, idx int, arguments []any) (any, error) {
			// get(arr, 2)
			return (*slice)[idx], nil
		}, func(map_ *map[string]any, key string, arguments []any) (any, error) {
			// get(arr, "hello")
			return (*map_)[key], nil
		},
	)
}

func builtinPush(env Env, e []Expression) (any, error) {
	// push(arr, 42)
	arr, err := e[0].evaluate(env)
	if err != nil {
		return nil, err
	}

	array, ok := arr.(*[]any)
	if !ok {
		return nil, fmt.Errorf("first argument needs to be an array, but was '%v'", arr)
	}

	if v, err := e[1].evaluate(env); err != nil {
		return nil, err
	} else {
		*array = append(*array, v)
		return v, nil
	}
}

func builtinPushVm(arguments []any) (any, error) {
	// push(arr, 42)
	arr := arguments[0]
	array, ok := arr.(*[]any)
	if !ok {
		return nil, fmt.Errorf("first argument needs to be an array, but was '%v'", arr)
	}

	v := arguments[1]
	*array = append(*array, v)
	return v, nil
}

func builtinSet(env Env, e []Expression) (any, error) {
	// set(arr, 2, 42) or set(map, "hello", 42)
	return arrayOrMapOp(env, e,
		func(slice *[]any, idx int, env Env, e []Expression) (any, error) {
			// set(arr, 2, 42)
			v, err := e[2].evaluate(env)
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
			v, err := e[2].evaluate(env)
			if err != nil {
				return nil, err
			}
			(*map_)[key] = v
			return v, nil
		},
	)
}

func builtinSetVm(arguments []any) (any, error) {
	// set(arr, 2, 42) or set(map, "hello", 42)
	return arrayOrMapOpVm(arguments,
		func(slice *[]any, idx int, arguments []any) (any, error) {
			// set(arr, 2, 42)
			v := arguments[2]
			if idx == len(*slice) {
				*slice = append(*slice, v)
			} else if idx < len(*slice) {
				(*slice)[idx] = v
			} else {
				return nil, fmt.Errorf("index %d out of bounds (length %d)", idx, len(*slice))
			}
			return v, nil
		}, func(map_ *map[string]any, key string, arguments []any) (any, error) {
			// set(arr, "hello", 42)
			v := arguments[2]
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

func builtinLenVm(arguments []any) (any, error) {
	// len(arr)
	slice, map_, err := getSliceOrMapVm(arguments)
	if err != nil {
		return nil, err
	} else if slice != nil {
		return len(*slice), nil
	} else {
		return len(*map_), nil
	}
}

func builtinKeys(env Env, e []Expression) (any, error) {
	// keys(arr)
	v, err := e[0].evaluate(env)
	if err != nil {
		return nil, err
	}

	map_, ok := v.(*map[string]any)
	if !ok {
		return nil, fmt.Errorf("argument to keys needs to be a map, but was '%v'", v)
	}

	return keys(map_), nil
}

func builtinKeysVm(arguments []any) (any, error) {
	// keys(arr)
	map_, err := getMapVm(arguments)
	if err != nil {
		return nil, err
	}

	return keys(map_), nil
}

func keys(map_ *map[string]any) *[]any {
	keys := make([]string, 0)
	for key := range *map_ {
		keys = append(keys, key)
	}
	// Sorting only to get consistent test results between invocations
	slices.Sort(keys)
	return toToiArray(keys)
}

func getMap(e []Expression, env Env) (*map[string]any, error) {
	v, err := e[0].evaluate(env)
	if err != nil {
		return nil, err
	}

	map_, ok := v.(*map[string]any)
	if ok {
		return map_, nil
	}
	return nil, fmt.Errorf("argument needs to be a map, but was '%v'", v)
}

func getMapVm(arguments []any) (*map[string]any, error) {
	v := arguments[0]

	map_, ok := v.(*map[string]any)
	if ok {
		return map_, nil
	}
	return nil, fmt.Errorf("argument needs to be a map, but was '%v'", v)
}

func toToiArray[T any](l []T) *[]any {
	array := make([]any, len(l))
	for i, s := range l {
		array[i] = s
	}
	return &array
}
