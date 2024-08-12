package main

import "fmt"

func decompile(ops []byte) {
	fmt.Println("Constants:")
	for i, constantValue := range constants {
		fmt.Printf("    %d: %v\n", i, constantValue)
	}

	fmt.Println("\nOps:")
	i := 0
	for i < len(ops) {
		op := ops[i]
		i++

		fmt.Printf("    %d: (%d) ", i, op)
		switch op {
		case OpPop:
			fmt.Print("Pop")
		case OpBinary:
			fmt.Print("Binary")
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
			fmt.Print("Not")
		case OpPrintln:
			fmt.Print("PrintLn")
		case OpJumpIfTrue:
			num := ops[i]
			i++
			fmt.Printf("JumpIfTrue +%d", num)
		case OpJumpBack:
			num := ops[i]
			i++
			fmt.Printf("JumpBack -%d", num)
		case OpInlineNumber:
			num := ops[i]
			i++
			fmt.Printf("InlineNumber %d", num)
		case OpLoadConstant:
			index := ops[i]
			i++
			constantValue := constants[index]
			fmt.Printf("LoadConstant %d '%v'", index, constantValue)
		case OpReadVariable:
			index := ops[i]
			constantValue := constants[index]
			i++
			fmt.Printf("ReadVariable %d '%v'", index, constantValue)
		case OpSetVariable:
			index := ops[i]
			constantValue := constants[index]
			i++
			fmt.Printf("SetVariable %d '%v'", index, constantValue)
		case OpCallBuiltin:
			index := ops[i]
			i++
			constantValue := constants[index]
			fmt.Printf("Builtin %d '%v'", index, constantValue)
		case OpDuplicate:
			fmt.Print("Duplicate")
		case InvalidOp:
			fmt.Print("!! Invalid op !!")
		}
		fmt.Println()
	}
}
