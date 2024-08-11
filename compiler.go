package main

import "fmt"

// TODO: no global state
var constants = make([]any, 0)

// TODO: use a bytebuffer instead of slices for efficiency; although slices are nice and easy to patch jumps

// Statements

func (s *BlockStatement) compile() ([]byte, error) {
	ops := make([]byte, 0)
	for _, stmt := range s.Statements {
		stmtOps, err := stmt.compile()
		if err != nil {
			return nil, err
		}
		ops = append(ops, stmtOps...)
	}
	return ops, nil
}

func (s *IfStatement) compile() ([]byte, error) {
	condition, err := s.Condition.compile()
	if err != nil {
		return nil, err
	}

	not := []byte{OpNot} // We don't have a "jump if false", so we need to NOT the result
	jump := []byte{OpJumpIfTrue, InvalidOp}

	then, err := s.Then.compile()
	if err != nil {
		return nil, err
	}

	if len(then) > MaxBlockSize {
		// TODO: keep tokens for error reporting
		return nil, fmt.Errorf("if's then block of %d statements exceeds %d operations (token %s; '%s')", len(then), MaxBlockSize, "oops", "oops")
	}

	// Patch jump now that we know the number of instructions to jump over
	jump[1] = byte(len(then))

	return combine(condition, not, jump, then), nil
}

func (s *WhileStatement) compile() ([]byte, error) {
	condition, err := s.Condition.compile()
	if err != nil {
		return nil, err
	}
	not := []byte{OpNot} // We don't have a "jump if false", so we need to NOT the result
	jump := []byte{OpJumpIfTrue, InvalidOp}

	body, err := s.Body.compile()
	if err != nil {
		return nil, err
	}

	if len(body) > MaxBlockSize {
		// TODO: keep tokens for error reporting
		return nil, fmt.Errorf("while body of %d statements exceeds %d operations (token %s; '%s')", len(body), MaxBlockSize, "oops", "oops")
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
		// TODO: keep tokens for error reporting
		return nil, fmt.Errorf("backjump of %d statements exceeds %d operations (token %s; '%s')", jumpBackCount, MaxBlockSize, "oops", "oops")
	}
	jumpBack := []byte{OpJumpBack, byte(jumpBackCount)}

	// Patch jump now that we know the number of instructions to jump over
	jump[1] = byte(len(body) + 2) // +2 to jump over the OpJumpBack instruction and its argument

	return combine(condition, not, jump, body, jumpBack), nil
}

func (s *AssignmentStatement) compile() ([]byte, error) {
	index, err := ensureConstant(s.Identifier.Lexeme)
	if err != nil {
		return nil, err
	}

	expression, err := s.Expression.compile()
	if err != nil {
		return nil, err
	}

	return combine(expression, []byte{OpSetVariable, index}), nil
}

func (s *ExpressionStatement) compile() ([]byte, error) {
	/* Discard return value afterwards using pop */
	ops, err := s.Expression.compile()
	if err != nil {
		return nil, err
	}
	return combine(ops, []byte{OpPop}), nil
}

// Expressions

func (e *BinaryExpression) compile() ([]byte, error) {
	leftOps, err := e.Left.compile()
	if err != nil {
		return nil, err
	}
	rightOps, err := e.Right.compile()
	if err != nil {
		return nil, err
	}

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
	return combine(leftOps, rightOps, ops), nil
}

func (e *FunctionCallExpression) compile() ([]byte, error) {
	if len(e.Arguments) > 50 {
		return nil, fmt.Errorf("functions don't support more than 50 arguments (was %d for '%v')", len(e.Arguments), e.Token.Lexeme)
	}

	ops := make([]byte, 0)
	for _, arg := range e.Arguments {
		exprOps, err := arg.compile()
		if err != nil {
			return nil, err
		}
		ops = append(ops, exprOps...)
	}

	if e.Token.Lexeme == "println" {
		return append(ops, []byte{OpPrintln, byte(len(e.Arguments))}...), nil
	}

	index, err := ensureConstant(e.Token.Lexeme)
	if err != nil {
		return nil, err
	}
	return append(ops, []byte{OpCallBuiltin, index}...), nil
}

func (e *LiteralExpression) compile() ([]byte, error) {
	if i, ok := e.Token.Literal.(int); ok && i <= 0xFF {
		return []byte{OpInlineNumber, byte(i)}, nil
	}

	index, err := ensureConstant(e.Token.Literal)
	if err != nil {
		return nil, err
	}
	return []byte{OpLoadConstant, index}, nil
}

func (e *VariableExpression) compile() ([]byte, error) {
	identifier := e.Token.Lexeme
	index, err := ensureConstant(identifier)
	if err != nil {
		return nil, err
	}

	return []byte{OpReadVariable, index}, nil
}

func combine(slices ...[]byte) []byte {
	target := slices[0]
	for i := 1; i < len(slices); i++ {
		target = append(target, slices[i]...)
	}
	return target
}

func ensureConstant(value any) (byte, error) {
	for i, v := range constants {
		if v == value {
			return byte(i), nil
		}
	}

	if len(constants) == MaxConstants {
		return 0, fmt.Errorf("cannot add constant '%v' because the maximum of %d was reached", value, MaxConstants)
	}

	constants = append(constants, value)
	return byte(len(constants) - 1), nil
}
