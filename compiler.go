package main

import "fmt"

// TODO: no global state
var constants = make([]any, 0)

// Statements

func (s *BlockStatement) compile() []byte {
	ops := make([]byte, 0)
	for _, stmt := range s.Statements {
		ops = append(ops, stmt.compile()...)
	}
	return ops
}

func (s *IfStatement) compile() []byte {
	condition := s.Condition.compile()
	not := []byte{OpNot} // We don't have a "jump if false", so we need to NOT the result
	jump := []byte{OpJumpIfTrue, InvalidOp}

	then := s.Then.compile()

	if len(then) > MaxBlockSize {
		// TODO: compile should return error
		// TODO: keep tokens for error reporting
		panic(fmt.Sprintf("if's then block of %d statements exceeds %d operations (token %s; '%s')", len(then), MaxBlockSize, "oops", "oops"))
	}

	// Patch jump now that we know the number of instructions to jump over
	jump[1] = byte(len(then))

	return combine(condition, not, jump, then)
}

func (s *WhileStatement) compile() []byte {
	condition := s.Condition.compile()
	not := []byte{OpNot} // We don't have a "jump if false", so we need to NOT the result
	jump := []byte{OpJumpIfTrue, InvalidOp}

	body := s.Body.compile()

	if len(body) > MaxBlockSize {
		// TODO: compile should return error
		// TODO: keep tokens for error reporting
		panic(fmt.Sprintf("while body of %d statements exceeds %d operations (token %s; '%s')", len(body), MaxBlockSize, "oops", "oops"))
	}

	/*
		Layout is:
			- N ops for condition expression
			- 1 op for NOT of condition expression
			- 2 (op+arg) for jump if true
			- N for body
	*/
	jumpBackCount := len(condition) + len(not) + len(jump) + len(body) + 2 // + 2 for jumpBack + count
	if jumpBackCount > MaxBlockSize {
		// TODO: compile should return error
		// TODO: keep tokens for error reporting
		panic(fmt.Sprintf("backjump of %d statements exceeds %d operations (token %s; '%s')", jumpBackCount, MaxBlockSize, "oops", "oops"))
	}
	jumpBack := []byte{OpJumpBack, byte(jumpBackCount)}

	// Patch jump now that we know the number of instructions to jump over
	jump[1] = byte(len(body) + 2) // +2 to jump over the OpJumpBack instruction and its argument

	return combine(condition, not, jump, body, jumpBack)
}

func (s *AssignmentStatement) compile() []byte {
	index := ensureConstant(s.Identifier.Lexeme)

	expression := s.Expression.compile()

	return combine(expression, []byte{OpSetVariable, index})
}

func (s *ExpressionStatement) compile() []byte {
	/* Discard return value afterwards using pop */
	return combine(s.Expression.compile(), []byte{OpPop})
}

// Expressions

func (e *BinaryExpression) compile() []byte {
	leftOps := e.Left.compile()
	rightOps := e.Right.compile()

	var binaryOp byte
	appendNot := false
	switch e.Operator.Type {
	case TokenPlus:
		binaryOp = OpBinaryPlus
	case TokenMinus:
		binaryOp = OpBinarySubtract
	case TokenAsterisk:
		binaryOp = OpBinaryMultiply
	case TokenSlash:
		binaryOp = OpBinaryDivide

	case TokenUnderscore:
		binaryOp = OpBinaryConcat

	case TokenEqualEqual:
		binaryOp = OpBinaryEqual
	case TokenNotEqual:
		binaryOp = OpBinaryEqual
		appendNot = true
	case TokenGreaterThan:
		binaryOp = OpBinaryGreaterThan
	case TokenGreaterEqual:
		binaryOp = OpBinaryLessThan
		appendNot = true
	case TokenLessThan:
		binaryOp = OpBinaryLessThan
	case TokenLessEqual:
		binaryOp = OpBinaryGreaterThan
		appendNot = true
	}

	ops := []byte{OpBinary, binaryOp}
	if appendNot {
		ops = append(ops, OpNot)
	}
	return combine(leftOps, rightOps, ops)
}

func (e *FunctionCallExpression) compile() []byte {
	if e.Token.Lexeme == "println" {
		// TODO: better
		ops := make([]byte, 0)
		for _, arg := range e.Arguments {
			ops = append(ops, arg.compile()...)
		}
		// TODO: handle byte overflow
		return append(ops, []byte{OpPrintln, byte(len(e.Arguments))}...)
	}

	index := ensureConstant(e.Token.Lexeme)
	ops := make([]byte, 0)
	for _, arg := range e.Arguments {
		ops = append(ops, arg.compile()...)
	}
	return append(ops, []byte{OpCallBuiltin, index}...)
}

func (e *LiteralExpression) compile() []byte {
	if i, ok := e.Token.Literal.(int); ok && i <= 0xFF {
		return []byte{OpInlineNumber, byte(i)}
	}

	index := ensureConstant(e.Token.Literal)
	return []byte{OpLoadConstant, index}
}

func (e *VariableExpression) compile() []byte {
	identifier := e.Token.Lexeme
	index := ensureConstant(identifier)

	return []byte{OpReadVariable, index}
}

func combine(slices ...[]byte) []byte {
	target := slices[0]
	for i := 1; i < len(slices); i++ {
		target = append(target, slices[i]...)
	}
	return target
}

func ensureConstant(value any) byte {
	for i, v := range constants {
		if v == value {
			// TODO: remove debug logging
			fmt.Printf("re-using id %d for duplicate constant '%v'\n", i, v)
			return byte(i)
		}
	}

	if len(constants) == MaxConstants {
		// TODO: return error
		panic(fmt.Sprintf("cannot add constant '%v' because the maximum of %d was reached", value, MaxConstants))
	}

	constants = append(constants, value)
	return byte(len(constants) - 1)
}
