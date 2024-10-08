package main

import "fmt"

func decompile(constants []any, ops []byte) {
	fmt.Println("Constants:")
	for i, constantValue := range constants {
		fmt.Printf("    %d: %v\n", i, constantValue)
	}

	fmt.Println("\nOps:")
	i := 0
	for i < len(ops) {
		op := ops[i]

		fmt.Printf("    %d: (%d) ", i, op)
		i++

		switch op {
		case OpPop:
			fmt.Print("[1] Pop")
		case OpBinary:
			fmt.Print("[2] Binary")
			binop := ops[i]
			i++
			switch binop {
			case OpBinaryPlus:
				fmt.Print(" Plus")
			case OpBinarySubtract:
				fmt.Print(" Subtract")
			case OpBinaryMultiply:
				fmt.Print(" Multiply")
			case OpBinaryDivide:
				fmt.Print(" Divide")
			case OpBinaryEqual:
				fmt.Print(" Equal")
			case OpBinaryGreaterThan:
				fmt.Print(" GreaterThan")
			case OpBinaryLessThan:
				fmt.Print(" LessThan")
			case OpBinaryConcat:
				fmt.Print(" Concat")
			}
		case OpNot:
			fmt.Print("[1] Not")
		case OpCallVariadicFunction:
			argCount := int(ops[i])
			i++
			fmt.Printf("[2] PrintLn of %d arguments", argCount)
		case OpJumpIfFalse:
			num1 := int(ops[i])
			i++
			num2 := int(ops[i])
			i++
			jumpAmount := num1*256 + num2
			fmt.Printf("[3] JumpIfFalse +%d -> %d", jumpAmount, i+jumpAmount)
		case OpJumpForward:
			num1 := int(ops[i])
			i++
			num2 := int(ops[i])
			i++
			jumpAmount := num1*256 + num2
			fmt.Printf("[3] JumpForward +%d -> %d", jumpAmount, i+jumpAmount)
		case OpJumpBack:
			num1 := int(ops[i])
			i++
			num2 := int(ops[i])
			i++
			jumpAmount := num1*256 + num2
			fmt.Printf("[3] JumpBack -%d -> %d", jumpAmount, i-jumpAmount)
		case OpInlineNumber:
			num := ops[i]
			i++
			fmt.Printf("[2] InlineNumber %d", num)
		case OpLoadConstant:
			index := ops[i]
			i++
			constantValue := constants[index]
			fmt.Printf("[2] LoadConstant %d '%v'", index, constantValue)
		case OpReadVariable:
			index := ops[i]
			constantValue := constants[index]
			i++
			fmt.Printf("[2] ReadVariable %d '%v'", index, constantValue)
		case OpSetVariable:
			index := ops[i]
			constantValue := constants[index]
			i++
			fmt.Printf("[2] SetVariable %d '%v'", index, constantValue)
		case OpInstantiate:
			index := ops[i]
			constantValue := constants[index]
			i++
			fmt.Printf("[2] Instantiate %d '%v'", index, constantValue)
		case OpCallBuiltin:
			index := ops[i]
			i++
			constantValue := constants[index]
			fmt.Printf("[2] Builtin call %d '%v'", index, constantValue)
		case OpCallFunction:
			index := ops[i]
			i++
			constantValue := constants[index]
			fmt.Printf("[2] Function call %d '%v'", index, constantValue)
		case OpFieldAccess:
			index := ops[i]
			constantValue := constants[index]
			i++
			fmt.Printf("[2] Field access %d '%v'", index, constantValue)
		case OpSetField:
			index := ops[i]
			constantValue := constants[index]
			i++
			fmt.Printf("[2] Set field %d '%v'", index, constantValue)
		case OpDuplicate:
			fmt.Print("[1] Duplicate")
		case InvalidOp:
			fmt.Print("[1] !! Invalid op !!")
		}
		fmt.Println()
	}
	fmt.Printf("    Exit position: %d\n", i)
}
