package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
)

func dump(filename string, ops []byte, constants []any, variableDefinitions []string, functions map[string]VmFunction) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	fmt.Fprintf(w, "%d constants\n", len(constants))
	for _, v := range constants {
		if num, ok := v.(int); ok {
			fmt.Fprintf(w, "int:%d\n", num)
		} else if str, ok := v.(string); ok {
			fmt.Fprintf(w, "string:%s\n", str)
		} else {
			return fmt.Errorf("unsupported constant type %v for '%v'", reflect.TypeOf(v), v)
		}
	}

	fmt.Fprintf(w, "%d functions\n", len(functions))
	for name, f := range functions {
		fmt.Fprintf(w, "%s\n", name)
		fmt.Fprintf(w, "%t\n", f.hasOutVar)
		fmt.Fprintf(w, "%d parameters\n", len(f.params))
		for _, param := range f.params {
			fmt.Fprintf(w, "%s\n", param)
		}
		fmt.Fprintf(w, "%d variables\n", len(f.variableDefinitions))
		for _, name := range f.variableDefinitions {
			fmt.Fprintf(w, "%s\n", name)
		}
		fmt.Fprintf(w, "%d instructions\n", len(f.ops))
		for _, op := range f.ops {
			fmt.Fprintf(w, "%d\n", op)
		}
	}

	fmt.Fprintf(w, "%d variables\n", len(variableDefinitions))
	for _, name := range variableDefinitions {
		fmt.Fprintf(w, "%s\n", name)
	}

	fmt.Fprintf(w, "%d instructions\n", len(ops))
	for _, op := range ops {
		fmt.Fprintf(w, "%d\n", op)
	}

	return nil
}
