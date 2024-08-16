package main

import "fmt"

// TODO: use a bytebuffer instead of slices for efficiency; although slices are nice and easy to patch jumps

type Compiler struct {
	constants []any
}

// Statements

func (s *BlockStatement) compile(compiler *Compiler) ([]byte, error) {
	ops := make([]byte, 0)
	for _, stmt := range s.Statements {
		stmtOps, err := stmt.compile(compiler)
		if err != nil {
			return nil, err
		}
		ops = append(ops, stmtOps...)
	}
	return ops, nil
}

func (s *IfStatement) compile(compiler *Compiler) ([]byte, error) {
	condition, err := s.Condition.compile(compiler)
	if err != nil {
		return nil, err
	}

	then, err := s.Then.compile(compiler)
	if err != nil {
		return nil, err
	}

	var otherwise []byte
	if s.Otherwise != nil {
		otherwise, err = (*s.Otherwise).compile(compiler)
		if err != nil {
			return nil, err
		}
	}

	var jumpOverOtherwise []byte
	if len(otherwise) != 0 {
		b1, b2 := encodeJumpAmount(len(otherwise))
		jumpOverOtherwise = []byte{OpJumpForward, b1, b2}
	}

	jumpAmount := len(then) + len(jumpOverOtherwise)
	if jumpAmount > MaxBlockSize {
		// TODO: keep tokens for error reporting
		return nil, fmt.Errorf("if's then block of %d statements exceeds %d operations (token %s; '%s')", jumpAmount, MaxBlockSize, "oops", "oops")
	}

	b1, b2 := encodeJumpAmount(jumpAmount)
	jump := []byte{OpJumpIfFalse, b1, b2}

	return combine(condition, jump, then, jumpOverOtherwise, otherwise), nil
}

func (s *WhileStatement) compile(compiler *Compiler) ([]byte, error) {
	isForLoop := s.Token.Type == TokenFor
	condition, err := s.Condition.compile(compiler)
	if err != nil {
		return nil, err
	}
	jumpOutOfLoop := []byte{OpJumpIfFalse, InvalidOp, InvalidOp}

	body, err := s.Body.compile(compiler)
	if err != nil {
		return nil, err
	}
	if s.AfterBody != nil {
		afterBody, err := s.AfterBody.compile(compiler)
		if err != nil {
			return nil, err
		}
		body = append(body, afterBody...)
	}

	/*
		Layout is:
			- N ops for condition expression
			- 1 op for NOT of condition expression
			- 2 (op+arg) for jump if true
			- N for body
	*/
	jumpBackAmount := len(condition) + len(jumpOutOfLoop) + len(body) + 3 // + 3 for jumpBack + count
	if jumpBackAmount > MaxBlockSize {
		// TODO: keep tokens for error reporting
		return nil, fmt.Errorf("backjump of %d statements exceeds %d operations (token %s; '%s')", jumpBackAmount, MaxBlockSize, "oops", "oops")
	}
	b1, b2 := encodeJumpAmount(jumpBackAmount)
	jumpBack := []byte{OpJumpBack, b1, b2}

	// Jump forward
	jumpOutAmount := len(body) + 3 // +3 to jump over the OpJumpBack instruction and its argument
	b1, b2 = encodeJumpAmount(jumpOutAmount)
	jumpOutOfLoop[1] = b1
	jumpOutOfLoop[2] = b2

	// Search for "next iteration" statements and patch their jumps
	for i := 0; i < len(body)-2; i += 1 {
		// TODO: this is some ugly nasty ass shit that should never be allowed
		if body[i] != OpCompileTimeOnlyNextIteration || body[i+1] != InvalidOp || body[i+2] != InvalidOp {
			continue
		}

		body[i] = OpJumpForward
		// Distance between this statement and end of loop is jumpAmount - i
		// I don't know why we need '- 3' and I'm too tired to figure it out, but it works
		// In a "for" loop we have to jump to the index increment code; in a while loop, we can jumpt right out
		jumpAmount := jumpOutAmount - i - 3 - 3
		if isForLoop {
			jumpAmount -= 8
		}
		body[i+1], body[i+2] = encodeJumpAmount(jumpAmount)
	}

	// Search for "exit loop" statements and patch their jumps
	for i := 0; i < len(body)-2; i += 1 {
		// TODO: this is some ugly nasty ass shit that should never be allowed
		if body[i] != OpCompileTimeOnlyExitLoop || body[i+1] != InvalidOp || body[i+2] != InvalidOp {
			continue
		}

		body[i] = OpJumpForward
		// Distance between this statement and end of loop is jumpAmount - i
		// I don't know why we need '- 3' and I'm too tired to figure it out, but it works
		jumpAmount := jumpOutAmount - i - 3
		body[i+1], body[i+2] = encodeJumpAmount(jumpAmount)
	}

	return combine(condition, jumpOutOfLoop, body, jumpBack), nil
}

func (s *ExitLoopStatement) compile(compiler *Compiler) ([]byte, error) {
	// OpCompileTimeOnlyExitLoop will be replaced with a jump in the compile() function of While
	return []byte{OpCompileTimeOnlyExitLoop, InvalidOp, InvalidOp}, nil
}

func (s *NextIterationStatement) compile(compiler *Compiler) ([]byte, error) {
	// OpCompileTimeOnlyNextIteration will be replaced with a jump in the compile() function of While
	return []byte{OpCompileTimeOnlyNextIteration, InvalidOp, InvalidOp}, nil
}

func encodeJumpAmount(amount int) (byte, byte) {
	return byte(amount / 256), byte(amount % 256)
}

func (s *AssignmentStatement) compile(compiler *Compiler) ([]byte, error) {
	index, err := compiler.ensureConstant(s.Identifier.Lexeme)
	if err != nil {
		return nil, err
	}

	expression, err := s.Expression.compile(compiler)
	if err != nil {
		return nil, err
	}

	return combine(expression, []byte{OpSetVariable, index}), nil
}

func (s *ExpressionStatement) compile(compiler *Compiler) ([]byte, error) {
	/* Discard return value afterwards using pop */
	ops, err := s.Expression.compile(compiler)
	if err != nil {
		return nil, err
	}
	return combine(ops, []byte{OpPop}), nil
}

// Expressions

func (e *BinaryExpression) compile(compiler *Compiler) ([]byte, error) {
	if e.Operator.Type == TokenPipe {
		return e.compileOrOrAnd(compiler, true)
	} else if e.Operator.Type == TokenAmpersand {
		return e.compileOrOrAnd(compiler, false)
	}

	leftOps, err := e.Left.compile(compiler)
	if err != nil {
		return nil, err
	}
	rightOps, err := e.Right.compile(compiler)
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
	default:
		return nil, fmt.Errorf("unsupported binary operator %v ('%v')", e.Operator.Type, e.Operator.Lexeme)
	}

	ops := []byte{OpBinary, binaryOp}
	if appendNot {
		ops = append(ops, OpNot)
	}
	return combine(leftOps, rightOps, ops), nil
}

func (e *BinaryExpression) compileOrOrAnd(compiler *Compiler, withNot bool) ([]byte, error) {
	leftOps, err := e.Left.compile(compiler)
	if err != nil {
		return nil, err
	}

	rightOps, err := e.Right.compile(compiler)
	if err != nil {
		return nil, err
	}

	jumpAmount := len(rightOps) + 1 // include the Pop
	b1, b2 := encodeJumpAmount(jumpAmount)

	dupe := []byte{OpDuplicate}
	not := []byte{}
	if withNot {
		not = []byte{OpNot}
	}
	jump := []byte{OpJumpIfFalse, b1, b2, OpPop}

	return combine(leftOps, dupe, not, jump, rightOps), nil
}

func (e *ContainerAccessExpression) compile(compiler *Compiler) ([]byte, error) {
	f := &FunctionCallExpression{
		Token:        e.Token,
		FunctionName: "get",
		Arguments:    []Expression{e.Container, e.Access},
	}
	return f.compile(compiler)
}

func (e *FunctionCallExpression) compile(compiler *Compiler) ([]byte, error) {
	if len(e.Arguments) > 50 {
		return nil, fmt.Errorf("functions don't support more than 50 arguments (was %d for '%v')", len(e.Arguments), e.FunctionName)
	}

	ops := make([]byte, 0)
	for _, arg := range e.Arguments {
		exprOps, err := arg.compile(compiler)
		if err != nil {
			return nil, err
		}
		ops = append(ops, exprOps...)
	}

	if e.FunctionName == "println" {
		return append(ops, []byte{OpPrintln, byte(len(e.Arguments))}...), nil
	}

	index, err := compiler.ensureConstant(e.FunctionName)
	if err != nil {
		return nil, err
	}
	return append(ops, []byte{OpCallBuiltin, index}...), nil
}

func (e *LiteralExpression) compile(compiler *Compiler) ([]byte, error) {
	if i, ok := e.Token.Literal.(int); ok && i <= 0xFF {
		return []byte{OpInlineNumber, byte(i)}, nil
	}

	index, err := compiler.ensureConstant(e.Token.Literal)
	if err != nil {
		return nil, err
	}
	return []byte{OpLoadConstant, index}, nil
}

func (e *VariableExpression) compile(compiler *Compiler) ([]byte, error) {
	identifier := e.Token.Lexeme
	index, err := compiler.ensureConstant(identifier)
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

func (c *Compiler) ensureConstant(value any) (byte, error) {
	for i, v := range c.constants {
		if v == value {
			return byte(i), nil
		}
	}

	if len(c.constants) == MaxConstants {
		return 0, fmt.Errorf("cannot add constant '%v' because the maximum of %d was reached", value, MaxConstants)
	}

	c.constants = append(c.constants, value)
	return byte(len(c.constants) - 1), nil
}
