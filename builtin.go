package main

import (
	"bytes"
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
	"println":    {ArityVariadic, builtinPrintln, builtinPrintlnVm},
	"inputLines": {0, builtinInputLines, builtinInputLinesVm},

	"split": {2, builtinSplit, builtinSplitVm},
	"chars": {1, builtinChars, builtinCharsVm},

	"int":    {1, builtinInt, builtinIntVm},
	"string": {1, builtinString, builtinStringVm},

	// "Arrays" and "Maps"
	"array": {ArityVariadic, builtinArray, builtinArrayVm},
	"map":   {0, builtinMap, builtinMapVm},
	"get":   {2, builtinGet, builtinGetVm},
	"push":  {2, builtinPush, builtinPushVm},
	"pop":   {1, builtinPop, builtinPopVm},
	"set":   {3, builtinSet, builtinSetVm},
	"len":   {1, builtinLen, builtinLenVm},
	"keys":  {1, builtinKeys, builtinKeysVm},
	"isSet": {2, builtinIsSet, builtinIsSetVm},
	"unset": {2, builtinUnset, builtinUnsetVm},
}

func toArguments(env Env, e []Expression) ([]any, error) {
	arguments := make([]any, len(e))
	var err error
	for i, expr := range e {
		arguments[i], err = expr.evaluate(env)
		if err != nil {
			return nil, err
		}
	}
	return arguments, nil
}

func builtinPrintln(env Env, e []Expression) (any, error) {
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinPrintlnVm(arguments)
}

func builtinPrintlnVm(arguments []any) (any, error) {
	for i, v := range arguments {
		if i != 0 {
			toiStdout.WriteString(", ")
		}
		writeValue(v, toiStdout)
	}
	toiStdout.WriteRune('\n')
	return nil, nil
}

type printer interface {
	print(out *bytes.Buffer)
}

func writeValue(v any, out *bytes.Buffer) {
	if array, ok := v.(*[]any); ok {
		writeArray(array, out)
	} else if map_, ok := v.(*map[string]any); ok {
		writeMap(map_, out)
	} else if instance, ok := v.(printer); ok {
		instance.print(out)
	} else {
		out.WriteString(fmt.Sprintf("%v", v))
	}
}

func writeArray(array *[]any, out *bytes.Buffer) {
	out.WriteRune('[')
	for i, element := range *array {
		if i != 0 {
			out.WriteString(", ")
		}
		writeValue(element, out)
	}
	out.WriteRune(']')
}

func writeMap(map_ *map[string]any, out *bytes.Buffer) {
	out.WriteRune('{')
	for i, key := range sortedMapKeys(map_) {
		if i != 0 {
			out.WriteString(", ")
		}
		out.WriteString(fmt.Sprintf("%v", key))
		out.WriteString(": ")
		writeValue((*map_)[key], out)
	}
	out.WriteRune('}')
}

func builtinInputLines(env Env, e []Expression) (any, error) {
	return toToiArray(strings.Split(strings.TrimSpace(toiStdin), "\n")), nil
}

func builtinInputLinesVm(arguments []any) (any, error) {
	return toToiArray(strings.Split(strings.TrimSpace(toiStdin), "\n")), nil
}

func builtinSplit(env Env, e []Expression) (any, error) {
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinSplitVm(arguments)
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
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinCharsVm(arguments)
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
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinStringVm(arguments)
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
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinIntVm(arguments)
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
	arguments := make([]any, len(e))
	for i, expr := range e {
		value, err := expr.evaluate(env)
		if err != nil {
			return nil, err
		}
		arguments[i] = value
	}

	return builtinArrayVm(arguments)
}

func builtinArrayVm(arguments []any) (any, error) {
	return &arguments, nil
}

func builtinMap(env Env, e []Expression) (any, error) {
	return builtinMapVm([]any{})
}

func builtinMapVm(arguments []any) (any, error) {
	return &map[string]any{}, nil
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

func getArrayIndexVm(v any) (int, error) {
	if i, ok := v.(int); ok {
		return i, nil
	} else {
		return 0, fmt.Errorf("second argument needs to be a number, but was '%v'", v)
	}
}

func getMapKeyVm(v any) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	} else {
		return "", fmt.Errorf("second argument needs to be a string, but was '%v'", v)
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
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinGetVm(arguments)
}

func builtinGetVm(arguments []any) (any, error) {
	// get(arr, 2) or get(arr, "hello")
	return arrayOrMapOpVm(arguments,
		func(slice *[]any, idx int, arguments []any) (any, error) {
			// get(arr, 2)
			s := *slice
			if idx >= len(s) {
				return nil, fmt.Errorf("index out of bounds (requested %d; length %d)", idx, len(s))
			}
			return s[idx], nil
		}, func(map_ *map[string]any, key string, arguments []any) (any, error) {
			// get(arr, "hello")
			return (*map_)[key], nil
		},
	)
}

func builtinPush(env Env, e []Expression) (any, error) {
	// push(arr, 42)
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinPushVm(arguments)
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

func builtinPop(env Env, e []Expression) (any, error) {
	// pop(arr)
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinPopVm(arguments)
}

func builtinPopVm(arguments []any) (any, error) {
	// pop(arr)
	arr := arguments[0]
	array, ok := arr.(*[]any)
	if !ok {
		return nil, fmt.Errorf("first argument needs to be an array, but was '%v'", arr)
	}

	last := len(*array) - 1
	value := (*array)[last]
	*array = (*array)[:last]
	return value, nil
}

func builtinSet(env Env, e []Expression) (any, error) {
	// set(arr, 2, 42) or set(map, "hello", 42)
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinSetVm(arguments)
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
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinLenVm(arguments)
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
	// keys(map)
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinKeysVm(arguments)
}

func builtinKeysVm(arguments []any) (any, error) {
	// keys(map)
	slice, map_, err := getSliceOrMapVm(arguments)
	if err != nil {
		return nil, err
	} else if slice != nil {
		return indexes(slice), nil
	} else {
		return keys(map_), nil
	}
}

func indexes(array *[]any) *[]any {
	indexes := make([]int, len(*array))
	for i := range *array {
		indexes[i] = i
	}
	return toToiArray(indexes)
}

func keys(map_ *map[string]any) *[]any {
	return toToiArray(sortedMapKeys(map_))
}

func sortedMapKeys(map_ *map[string]any) []string {
	keys := make([]string, 0)
	for key := range *map_ {
		keys = append(keys, key)
	}
	// Sorting only to get consistent test results between invocations
	slices.Sort(keys)
	return keys
}

func builtinIsSet(env Env, e []Expression) (any, error) {
	// isSet(map, "key")
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinIsSetVm(arguments)
}

func builtinIsSetVm(arguments []any) (any, error) {
	// isSet(map, "key")
	map_, err := getMapVm(arguments)
	if err != nil {
		return nil, err
	}

	key, err := getMapKeyVm(arguments[1])
	if err != nil {
		return nil, err
	}

	if _, found := (*map_)[key]; found {
		return 1, nil
	} else {
		return 0, nil
	}
}

func builtinUnset(env Env, e []Expression) (any, error) {
	// unset(map, "key")
	arguments, err := toArguments(env, e)
	if err != nil {
		return nil, err
	}
	return builtinUnsetVm(arguments)
}

func builtinUnsetVm(arguments []any) (any, error) {
	// isSet(map, "key")
	map_, err := getMapVm(arguments)
	if err != nil {
		return nil, err
	}

	key, err := getMapKeyVm(arguments[1])
	if err != nil {
		return nil, err
	}

	delete(*map_, key)
	return 0, nil
}

func getMapVm(arguments []any) (*map[string]any, error) {
	v := arguments[0]

	map_, ok := v.(*map[string]any)
	if ok {
		return map_, nil
	}
	return nil, fmt.Errorf("first argument needs to be a map, but was '%v'", v)
}

func toToiArray[T any](l []T) *[]any {
	array := make([]any, len(l))
	for i, s := range l {
		array[i] = s
	}
	return &array
}
