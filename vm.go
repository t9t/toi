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
	OpJumpIfFalse
	OpJumpForward
	OpJumpBack
	OpInlineNumber
	OpLoadConstant
	OpReadVariable
	OpSetVariable
	OpCallBuiltin
	OpCallFunction
	OpPrintln // Special op because it's variadic
	OpDuplicate

	InvalidOp
)

const (
	MaxBlockSize = 60_000
	MaxConstants = 250
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

type Vm struct {
	ops       []byte
	constants []any
	globals   []any
	functions map[string][]byte
}

func execute(ops []byte, constants []any, functions map[string][]byte) error {
	globals := make([]any, len(constants)) // TODO: some memory is wasted here; not every constant needs a global
	vm := &Vm{
		ops:       ops,
		constants: constants,
		functions: functions,
		globals:   globals,
	}
	_, err := vm.execute()
	return err
}

func (vm *Vm) execute() ([]any, error) {
	constants, ops, functions := vm.constants, vm.ops, vm.functions

	ip := 0
	readOpByte := func() byte {
		op := ops[ip]
		ip++
		return op
	}
	getConstant := func(index int) (string, error) {
		constant := constants[index]
		constantString, ok := constant.(string)
		if !ok {
			return "", fmt.Errorf("expected constant %d to be a string, but was '%v' at %d", index, constant, ip)
		}
		return constantString, nil
	}
	readConstantString := func() (string, error) {
		return getConstant(int(readOpByte()))
	}

	maxStack := 20
	stack := make([]any, maxStack)
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
				return nil, fmt.Errorf("unsupported binary operator %v at %d", binop, ip)
			}

			if err != nil {
				return nil, err
			}

			pushStack(result)
		case OpNot:
			v := popStack()
			i, ok := v.(int)
			if !ok {
				return nil, fmt.Errorf("operand of NOT operation must be int, but was '%v' at %d", v, ip)
			}
			pushStack(boolToInt(!intToBool(i)))
		case OpJumpIfFalse:
			b1 := int(readOpByte())
			b2 := int(readOpByte())
			jumpAmount := b1*256 + b2
			v := popStack()
			if !isWeirdlyTrue(v) {
				ip += jumpAmount
			}
		case OpJumpForward:
			b1 := int(readOpByte())
			b2 := int(readOpByte())
			jumpAmount := b1*256 + b2
			ip += jumpAmount
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
			index := (int(readOpByte()))
			value := vm.globals[index]
			if value == nil {
				variableName, err := readConstantString()
				if err != nil {
					return nil, err
				}
				return nil, fmt.Errorf("variable '%v' not defined at %d", variableName, ip)
			}
			pushStack(value)
		case OpSetVariable:
			index := int(readOpByte())
			vm.globals[index] = popStack()
		case OpCallBuiltin:
			functionName, err := readConstantString()
			if err != nil {
				return nil, err
			}
			builtin, found := builtins[functionName]
			if !found {
				return nil, fmt.Errorf("builtin function '%v' not found at %d", functionName, ip)
			}
			arguments := make([]any, builtin.Arity)
			for i := 0; i < builtin.Arity; i++ {
				arguments[i] = popStack()
			}
			slices.Reverse(arguments) // Arguments were pushed onto the stack in left-to-right order, so we read them right-to-left
			returnValue, err := builtin.VmFunc(arguments)
			if err != nil {
				return nil, err
			}
			pushStack(returnValue)
		case OpCallFunction:
			functionName, err := readConstantString()
			if err != nil {
				return nil, err
			}
			// TODO: implement passing arguments
			functionVm := &Vm{
				ops:       functions[functionName],
				constants: constants,
				functions: functions,
				globals:   vm.globals,
			}

			_, err = functionVm.execute()
			if err != nil {
				return nil, err
			}

			// TODO: implement return values
			pushStack(nil)
		case OpPrintln:
			argumentCount := int(readOpByte())
			arguments := make([]any, argumentCount)
			for i := 0; i < argumentCount; i++ {
				arguments[i] = popStack()
			}
			slices.Reverse(arguments) // Arguments were pushed onto the stack in left-to-right order, so we read them right-to-left
			returnValue, err := builtinPrintlnVm(arguments)
			if err != nil {
				return nil, err
			}
			pushStack(returnValue)
		case OpDuplicate:
			v := popStack()
			pushStack(v)
			pushStack(v)

		default:
			return nil, fmt.Errorf("unknown instruction %v at %d", instruction, ip)
		}
	}

	fmt.Printf("VM run time: %v\n", time.Since(start))
	return stack, nil
}
