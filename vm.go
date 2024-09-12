package main

import (
	"bytes"
	"fmt"
	"slices"
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
	OpInstantiate
	OpCallBuiltin
	OpCallFunction
	OpPrintln // Special op because it's variadic
	OpFieldAccess
	OpSetField
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
	OpBinaryRemainder

	OpBinaryEqual
	OpBinaryGreaterThan
	OpBinaryLessThan

	OpBinaryBinaryOr
	OpBinaryBinaryXor
	OpBinaryBinaryAnd

	OpBinaryConcat
)

type VmType struct {
	Name   string
	Fields []string
}

type VmInstance struct {
	vmType *VmType
	values []any
}

func (instance *VmInstance) print(out *bytes.Buffer) {
	out.WriteString(instance.vmType.Name)
	out.WriteRune('{')
	for i := range len(instance.values) {
		fieldName := instance.vmType.Fields[i]
		fieldValue := instance.values[i]
		if i != 0 {
			out.WriteRune(',')
		}
		out.WriteString(fieldName)
		out.WriteRune('=')
		writeValue(fieldValue, out)
	}
	out.WriteRune('}')
}

type VmFunction struct {
	params              []string
	ops                 []byte
	variableDefinitions []string
	hasOutVar           bool
}

type Vm struct {
	ops                 []byte
	constants           []any
	variableDefinitions []string
	variables           []any
	functions           map[string]VmFunction
	types               map[string]VmType
}

const maxStack = 50

func execute(ops []byte, constants []any, variableDefinitions []string, functions map[string]VmFunction, types map[string]VmType) error {
	variables := make([]any, len(variableDefinitions))
	vm := &Vm{
		ops:                 ops,
		constants:           constants,
		functions:           functions,
		types:               types,
		variableDefinitions: variableDefinitions,
		variables:           variables,
	}
	stack := make([]any, maxStack)
	err := vm.execute(stack)
	return err
}

func (vm *Vm) execute(stack []any) error {
	constants, ops, functions, types := vm.constants, vm.ops, vm.functions, vm.types

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
			case OpBinaryRemainder:
				result, err = intBinaryOp(left, right, "%", func(l int, r int) int { return l % r })
			case OpBinaryBinaryAnd:
				result, err = intBinaryOp(left, right, "%", func(l int, r int) int { return l & r })
			case OpBinaryBinaryOr:
				result, err = intBinaryOp(left, right, "%", func(l int, r int) int { return l | r })
			case OpBinaryBinaryXor:
				result, err = intBinaryOp(left, right, "%", func(l int, r int) int { return l ^ r })

			case OpBinaryEqual:
				result = boolToInt(isEqual(left, right))
			case OpBinaryGreaterThan:
				result, err = intBinaryOp(left, right, ">", func(l int, r int) int { return boolToInt(l > r) })
			case OpBinaryLessThan:
				result, err = intBinaryOp(left, right, "<", func(l int, r int) int { return boolToInt(l < r) })

			case OpBinaryConcat:
				result, err = stringConcat(left, right)

			default:
				return fmt.Errorf("unsupported binary operator %v at %d", binop, ip)
			}

			if err != nil {
				return err
			}

			pushStack(result)
		case OpNot:
			v := popStack()
			i, ok := v.(int)
			if !ok {
				return fmt.Errorf("operand of NOT operation must be int, but was '%v' at %d", v, ip)
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
			index := int(readOpByte())
			value := vm.variables[index]
			if value == nil {
				variableName := vm.variableDefinitions[index]
				return fmt.Errorf("variable '%v' not defined at %d", variableName, ip)
			}
			pushStack(value)
		case OpSetVariable:
			index := int(readOpByte())
			vm.variables[index] = popStack()
		case OpInstantiate:
			typeName, err := readConstantString()
			if err != nil {
				return err
			}
			vmType, found := types[typeName]
			if !found {
				return fmt.Errorf("type '%v' not found at %d", typeName, ip)
			}
			fieldValues := make([]any, len(vmType.Fields))
			for i := range vmType.Fields {
				fieldValues[i] = popStack()
			}
			slices.Reverse(fieldValues) // Arguments were pushed onto the stack in left-to-right order, so we read them right-to-left
			instance := VmInstance{
				vmType: &vmType,
				values: fieldValues,
			}
			pushStack(&instance)
		case OpCallBuiltin:
			functionName, err := readConstantString()
			if err != nil {
				return err
			}
			builtin, found := builtins[functionName]
			if !found {
				return fmt.Errorf("builtin function '%v' not found at %d", functionName, ip)
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
		case OpCallFunction:
			functionName, err := readConstantString()
			if err != nil {
				return err
			}
			function := functions[functionName]
			functionVariables := make([]any, len(function.variableDefinitions))
			for i := len(function.params) - 1; i >= 0; i-- {
				functionVariables[i] = popStack()
			}

			functionVm := &Vm{
				ops:                 function.ops,
				constants:           constants,
				functions:           functions,
				variables:           functionVariables,
				variableDefinitions: function.variableDefinitions,
			}

			err = functionVm.execute(stack[stackNext:])
			if err != nil {
				return err
			}

			var outVar any = nil
			if function.hasOutVar {
				// E.g. if a function has 2 input parameters, and 1 output parameter, then the variable spot for the
				// output parameter is right after the input parameters, i.e. in the 3rd spot, or index 2, which is the
				// length of the params slice
				outVar = functionVariables[len(function.params)]
			}

			pushStack(outVar)
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
		case OpFieldAccess:
			identifier, err := readConstantString()
			if err != nil {
				return err
			}
			target := popStack()
			instance, ok := target.(*VmInstance)
			if !ok {
				return fmt.Errorf("left-hand operand of '.' must be a type instance but was '%v'", target)
			}
			fieldFound := false
			for i, field := range instance.vmType.Fields {
				if field == identifier {
					pushStack(instance.values[i])
					fieldFound = true
					break
				}
			}
			if !fieldFound {
				return fmt.Errorf("field '%v' not found on type '%v'", identifier, instance.vmType.Name)
			}
		case OpSetField:
			identifier, err := readConstantString()
			if err != nil {
				return err
			}
			value := popStack()
			target := popStack()
			instance, ok := target.(*VmInstance)
			if !ok {
				return fmt.Errorf("left-hand operand of '.' must be a type instance but was '%v'", target)
			}
			fieldFound := false
			for i, field := range instance.vmType.Fields {
				if field == identifier {
					instance.values[i] = value
					fieldFound = true
					break
				}
			}
			if !fieldFound {
				return fmt.Errorf("field '%v' not found on type '%v'", identifier, instance.vmType.Name)
			}
		case OpDuplicate:
			v := popStack()
			pushStack(v)
			pushStack(v)

		default:
			return fmt.Errorf("unknown instruction %v at %d", instruction, ip)
		}
	}

	return nil
}
