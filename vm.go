package main

import (
	"fmt"
	"slices"
	"time"
)

// type Opcode byte
// type BinaryOpcode byte
// TODO: don't want to be casting `byte(opcode)` all the time

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
	OpDuplicate
)

const (
	InvalidOp    byte = 0xFF
	MaxBlockSize      = 60_000
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
	readOpByte := func() byte {
		op := ops[ip]
		ip++
		return op
	}
	readConstantString := func() (string, error) {
		index := int(readOpByte())
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
	stackNext := 0
	popStack := func() any {
		stackNext -= 1
		return stack[stackNext]
	}
	pushStack := func(v any) {
		if stackNext == maxStack {
			panic(fmt.Sprintf("stack overflow: attempting to push '%v' onto the stack with maximum size %d", v, maxStack))
		}
		stack[stackNext] = v
		stackNext += 1
	}

	start := time.Now()
	for ip < len(ops) {
		instruction := readOpByte()

		switch instruction {
		case OpPop:
			_ = popStack()
		case OpBinary:
			binop := readOpByte()
			right := popStack()
			left := popStack()

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

			default:
				return fmt.Errorf("unsupported binary operator %v", binop)
			}

			if err != nil {
				return err
			}

			pushStack(result)
		case OpNot:
			v := popStack()
			i, ok := v.(int)
			if !ok {
				return fmt.Errorf("operand of NOT operation must be int, but was '%v'", v)
			}
			pushStack(boolToInt(!intToBool(i)))
		case OpJumpIfTrue:
			b1 := int(readOpByte())
			b2 := int(readOpByte())
			jumpAmount := b1*256 + b2
			v := popStack()
			if isWeirdlyTrue(v) {
				ip += jumpAmount
			}
		case OpJumpBack:
			b1 := int(readOpByte())
			b2 := int(readOpByte())
			jumpAmount := b1*256 + b2
			ip -= jumpAmount
		case OpInlineNumber:
			v := int(readOpByte())
			pushStack(v)
		case OpLoadConstant:
			index := int(readOpByte())
			pushStack(constants[index])
		case OpReadVariable:
			variableName, err := readConstantString()
			if err != nil {
				return err
			}
			value, found := variables[variableName]
			if !found {
				return fmt.Errorf("variable '%v' not defined", variableName)
			}
			pushStack(value)
		case OpSetVariable:
			variableName, err := readConstantString()
			if err != nil {
				return err
			}
			variables[variableName] = popStack()
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
				arguments[i] = popStack()
			}
			slices.Reverse(arguments) // Arguments were pushed onto the stack in left-to-right order, so we read them right-to-left
			returnValue, err := builtin.VmFunc(arguments)
			if err != nil {
				return err
			}
			pushStack(returnValue)
		case OpPrintln:
			argumentCount := int(readOpByte())
			arguments := make([]any, argumentCount)
			for i := 0; i < argumentCount; i++ {
				arguments[i] = popStack()
			}
			slices.Reverse(arguments) // Arguments were pushed onto the stack in left-to-right order, so we read them right-to-left
			returnValue, err := builtinPrintlnVm(arguments)
			if err != nil {
				return err
			}
			pushStack(returnValue)
		case OpDuplicate:
			v := popStack()
			pushStack(v)
			pushStack(v)
		default:
			return fmt.Errorf("unknown instruction %v", instruction)
		}
	}

	fmt.Printf("VM run time: %v\n", time.Since(start))
	return nil
}
