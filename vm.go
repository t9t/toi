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
	readConstantString := func() (string, error) {
		index := int(readByte())
		constant := constants[index]
		constantString, ok := constant.(string)
		if !ok {
			return "", fmt.Errorf("expected constant %d to be a string, but was '%v'", index, constant)
		}
		return constantString, nil
	}

	maxStack := 20
	stack := make([]any, maxStack)
	variables := make(map[string]any, 0)
	stackPos := 0
	pop := func() any {
		v := stack[stackPos]
		stackPos--
		return v
	}
	push := func(v any) {
		if stackPos == maxStack {
			panic(fmt.Sprintf("stack overflow: attempting to push '%v' onto the stack with maximum size %d", v, maxStack))
		}
		stackPos++
		stack[stackPos] = v
	}

	start := time.Now()
	for ip < len(ops) {
		instruction := readByte()

		switch instruction {
		case OpPop:
			_ = pop()
		case OpBinary:
			binop := readByte()
			right := pop()
			left := pop()

			var result any
			var err error

			switch binop {
			case OpBinaryPlus:
				result, err = intBinaryOp(left, right, "+", func(l int, r int) int { return l + r })
			case OpBinarySubtract:
				result, err = intBinaryOp(left, right, "-", func(l int, r int) int { return l - r })
			case OpBinaryMultiply:
				result, err = intBinaryOp(left, right, "*", func(l int, r int) int { return l * r })
			case OpBinaryDivide:
				result, err = intBinaryOp(left, right, "/", func(l int, r int) int { return l / r })

			case OpBinaryEqual:
				result = boolToInt(left == right)
			case OpBinaryGreaterThan:
				result, err = intBinaryOp(left, right, ">", func(l int, r int) int { return boolToInt(l > r) })
			case OpBinaryLessThan:
				result, err = intBinaryOp(left, right, "<", func(l int, r int) int { return boolToInt(l < r) })

			case OpBinaryConcat:
				result, err = stringConcat(left, right)
			}

			if err != nil {
				return err
			}

			push(result)
		case OpNot:
			v := pop()
			i, ok := v.(int)
			if !ok {
				return fmt.Errorf("operand of not operation must be int, but was '%v'", v)
			}
			push(boolToInt(!intToBool(i)))
		case OpJumpIfTrue:
			jumpAmount := int(readByte())
			v := pop()
			if isWeirdlyTrue(v) {
				ip += jumpAmount
			}
		case OpJumpBack:
			jumpAmount := int(readByte())
			ip -= jumpAmount
		case OpInlineNumber:
			v := int(readByte())
			push(v)
		case OpLoadConstant:
			index := int(readByte())
			push(constants[index])
		case OpReadVariable:
			variableName, err := readConstantString()
			if err != nil {
				return err
			}
			value, found := variables[variableName]
			if !found {
				return fmt.Errorf("variable '%v' not defined", variableName)
			}
			push(value)
		case OpSetVariable:
			variableName, err := readConstantString()
			if err != nil {
				return err
			}
			variables[variableName] = pop()
		case OpCallBuiltin:
			functionName, err := readConstantString()
			if err != nil {
				return err
			}
			builtin, found := builtins[functionName]
			if !found {
				return fmt.Errorf("builtin function '%v' not found", functionName)
			}
			arguments := make([]any, builtin.Arity)
			for i := 0; i < builtin.Arity; i++ {
				arguments[i] = pop()
			}
			slices.Reverse(arguments) // Arguments were pushed onto the stack in left-to-right order, so we read them right-to-left
			returnValue, err := builtin.VmFunc(arguments)
			if err != nil {
				return err
			}
			push(returnValue)
		case OpPrintln:
			argumentCount := int(readByte())
			arguments := make([]any, argumentCount)
			for i := 0; i < argumentCount; i++ {
				arguments[i] = pop()
			}
			slices.Reverse(arguments) // Arguments were pushed onto the stack in left-to-right order, so we read them right-to-left
			returnValue, err := builtinPrintlnVm(arguments)
			if err != nil {
				return err
			}
			push(returnValue)
		}
	}

	fmt.Printf("VM run time: %v\n", time.Since(start))
	return nil
}
