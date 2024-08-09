package main

import (
	"fmt"
	"slices"
	"time"
)

// TODO: don't want to be casting `byte(opcode)` all the time
// type Opcode byte
// type BinaryOpcode byte

const (
	OpPop byte = iota
	OpBinary
	OpNot
	OpJumpIfTrue
	OpJumpBack
	OpInlineNumber
	OpLoadConstant
	OpReadVariable
	OpSetVariable
	OpCallBuiltin
	OpPrintln // Special op because it's variadic
)

const (
	InvalidOp    byte = 0xFF
	MaxBlockSize      = 250
	MaxConstants      = 255
)

const (
	OpBinaryPlus byte = iota
	OpBinarySubtract
	OpBinaryMultiply
	OpBinaryDivide

	OpBinaryEqual
	OpBinaryGreaterThan
	OpBinaryLessThan

	OpBinaryConcat
)

func execute(ops []byte) error {
	ip := 0
	readByte := func() byte {
		op := ops[ip]
		ip++
		return op
	}

	maxStack := 100
	stack := make([]any, maxStack)
	variables := make(map[string]any, 0)
	stackPos := 0
	pop := func() any {
		v := stack[stackPos]
		stackPos--
		return v
	}
	popInt := func() int {
		// TODO: proper error handling
		return pop().(int)
	}
	push := func(v any) {
		if stackPos == maxStack {
			panic("stack overflow")
		}
		stackPos++
		stack[stackPos] = v
	}

	start := time.Now()
	for ip < len(ops) {
		// TODO: end of program condition
		instruction := readByte()
		//fmt.Printf("instruction: %d", instruction)
		//fmt.Printf("; stack (%d): %+v\n", stackPos, stack[:stackPos+1])

		switch instruction {
		case OpPop:
			_ = pop()
			//fmt.Printf("  pop '%v'\n", v)
		case OpBinary:
			binop := readByte()
			//fmt.Printf("  binary op %d; values:\n", binop)
			right := pop()
			left := pop()
			//fmt.Printf("    left: '%v'\n    right: '%v'\n", left, right)

			var result any
			bop := ""
			// TODO: proper type checking
			switch binop {
			case OpBinaryPlus:
				result = left.(int) + right.(int)
				bop = "+"
			case OpBinarySubtract:
				result = left.(int) - right.(int)
				bop = "-"
			case OpBinaryMultiply:
				result = left.(int) * right.(int)
				bop = "*"
			case OpBinaryDivide:
				result = left.(int) / right.(int)
				bop = "/"

			case OpBinaryEqual:
				result = boolToInt(left == right)
				bop = "="
			case OpBinaryGreaterThan:
				result = boolToInt(left.(int) > right.(int))
				bop = ">"
			case OpBinaryLessThan:
				result = boolToInt(left.(int) < right.(int))
				bop = "<"

			case OpBinaryConcat:
				result = left.(string) + right.(string)
				bop = "_"
			}

			//fmt.Printf("    -> op: %v; result: '%v'\n", bop, result)
			func(any) {}(bop)
			push(result)
		case OpNot:
			v := popInt()
			p := boolToInt(!intToBool(v))
			//fmt.Printf("  not popped: '%v'; pushing: '%v'\n", v, p)
			push(p)
		case OpJumpIfTrue:
			jumpAmount := int(readByte())
			v := pop()
			//fmt.Printf("  jump if true amount: %d; v: '%v'\n", jumpAmount, v)
			if isWeirdlyTrue(v) {
				//fmt.Printf("    -> jumping ip %d + %d = %d\n", ip, jumpAmount, ip+jumpAmount)
				ip += jumpAmount
			} else {
				//fmt.Printf("    -> not jumping\n")
			}
		case OpJumpBack:
			jumpAmount := int(readByte())
			//fmt.Printf("  jumping back %d; from %d to %d\n", jumpAmount, ip, ip-jumpAmount)
			ip -= jumpAmount
		case OpInlineNumber:
			v := int(readByte())
			//fmt.Printf("  read and pushing inline number %d\n", v)
			push(v)
		case OpLoadConstant:
			index := int(readByte())
			constantValue := constants[index]
			//fmt.Printf("  loading and pushing constant %d ('%v')\n", index, constantValue)
			push(constantValue)
		case OpReadVariable:
			index := int(readByte())
			constantValue := constants[index]
			//fmt.Printf("  reading and pushing variable %d ('%v')\n", index, constantValue)
			value, found := variables[constantValue.(string)]
			if !found {
				// TODO: better error handling
				panic(fmt.Sprintf("variable '%v' not defined", constantValue))
			}
			//fmt.Printf("    -> value: '%v'\n", value)
			push(value)
		case OpSetVariable:
			index := int(readByte())
			constantValue := constants[index]
			//fmt.Printf("  storing variable %d ('%v')\n", index, constantValue)
			newValue := pop()
			//oldValue := variables[constantValue.(string)]
			//fmt.Printf("    -> changing value from '%v' to '%v'\n", oldValue, newValue)
			variables[constantValue.(string)] = newValue
		case OpCallBuiltin:
			index := int(readByte())
			constantValue := constants[index]
			if constantValue == "println" {
				// TODO: do better
				panic("println not supported by OpCallBuiltin")
			}
			//fmt.Printf("  calling builtin function %d ('%v')\n", index, constantValue)
			f, found := builtins[constantValue.(string)]
			if !found {
				// TODO: better error handling; or do we need it at all?
				panic(fmt.Sprintf("builtin function '%v' not found", constantValue))
			}
			arguments := make([]any, f.Arity)
			for i := 0; i < f.Arity; i++ {
				arguments[i] = pop()
			}
			slices.Reverse(arguments) // Arguments were pushed onto the stack in left-to-right order, so we read them right-to-left
			v, err := f.VmFunc(arguments)
			if err != nil {
				// TODO: don't panic
				panic(fmt.Sprintf("error calling builtin '%v': %v", constantValue, err))
			}
			push(v)
		case OpPrintln:
			argumentCount := int(readByte())
			//fmt.Printf("  calling println with %d arguments\n", argumentCount)
			arguments := make([]any, argumentCount)
			for i := 0; i < argumentCount; i++ {
				arguments[i] = pop()
			}
			slices.Reverse(arguments) // Arguments were pushed onto the stack in left-to-right order, so we read them right-to-left
			v, err := builtinPrintlnVm(arguments)
			if err != nil {
				// TODO: don't panic
				panic(fmt.Sprintf("error calling println: %v", err))
			}
			push(v)
		}
	}
	fmt.Printf("VM run time: %v\n", time.Since(start))
	return nil
}
