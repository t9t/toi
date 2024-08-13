package main

import "fmt"

// Statements

func (s *BlockStatement) execute(env Env) error {
	for _, stmt := range s.Statements {
		if err := stmt.execute(env); err != nil {
			return err
		}
	}
	return nil
}

func (s *IfStatement) execute(env Env) error {
	v, err := s.Condition.evaluate(env)
	if err != nil {
		return err
	}
	if isWeirdlyTrue(v) {
		return s.Then.execute(env)
	} else if s.Otherwise != nil {
		return (*s.Otherwise).execute(env)
	}
	return nil
}

func (s *WhileStatement) execute(env Env) error {
	for {
		v, err := s.Condition.evaluate(env)
		if err != nil {
			return err
		}
		if !isWeirdlyTrue(v) {
			break
		}

		if err := s.Body.execute(env); err != nil {
			return err
		}
	}
	return nil
}

func (s *AssignmentStatement) execute(env Env) error {
	v, err := s.Expression.evaluate(env)
	if err != nil {
		return err
	}
	env[s.Identifier.Lexeme] = v
	return nil
}

func (s *ExpressionStatement) execute(env Env) error {
	_, err := s.Expression.evaluate(env) /* Discard return value */
	return err
}

// Expressions

func (e *BinaryExpression) evaluate(env Env) (any, error) {
	if e.Operator.Type == TokenAmpersand {
		return e.evaluateAnd(env)
	}

	left, err := e.Left.evaluate(env)
	if err != nil {
		return nil, err
	}

	right, err := e.Right.evaluate(env)
	if err != nil {
		return nil, err
	}

	token := e.Operator
	operator := token.Lexeme

	switch e.Operator.Type {
	case TokenPlus:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return l + r })
	case TokenMinus:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return l - r })
	case TokenAsterisk:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return l * r })
	case TokenSlash:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return l / r })

	case TokenUnderscore:
		return stringConcat(left, right)

	case TokenEqualEqual:
		return boolToInt(left == right), nil
	case TokenNotEqual:
		return boolToInt(left != right), nil
	case TokenGreaterThan:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return boolToInt(l > r) })
	case TokenGreaterEqual:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return boolToInt(l >= r) })
	case TokenLessThan:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return boolToInt(l < r) })
	case TokenLessEqual:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return boolToInt(l <= r) })
	}

	return nil, fmt.Errorf("unsupported binary operator %v ('%v')", token.Type, token.Lexeme)
}

func (e *BinaryExpression) evaluateAnd(env Env) (any, error) {
	left, err := e.Left.evaluate(env)
	if err != nil {
		return nil, err
	}

	leftInt, err := castToInt(left, "left", e.Operator.Lexeme)
	if err != nil {
		return nil, err
	}

	if !isWeirdlyTrue(leftInt) {
		return leftInt, nil
	}

	right, err := e.Right.evaluate(env)
	if err != nil {
		return nil, err
	}

	rightInt, err := castToInt(right, "right", e.Operator.Lexeme)
	if err != nil {
		return nil, err
	}

	return rightInt, nil
}

func intBinaryOp(left, right any, operator string, op func(int, int) int) (any, error) {
	leftInt, err := castToInt(left, "left", operator)
	if err != nil {
		return nil, err
	}

	rightInt, err := castToInt(right, "right", operator)
	if err != nil {
		return nil, err
	}

	return op(leftInt, rightInt), nil
}

func castToInt(v any, side, operator string) (int, error) {
	int, ok := v.(int)
	if !ok {
		return 0, fmt.Errorf("%s-hand operand of '%s' should be an int but was '%v'", side, operator, v)
	}
	return int, nil
}

func stringConcat(left, right any) (any, error) {
	leftString, ok := left.(string)
	if !ok {
		return nil, fmt.Errorf("left-hand operand of '_' should be a string but was '%v'", left)
	}

	rightString, ok := right.(string)
	if !ok {
		return nil, fmt.Errorf("right-hand operand of '_' should be a string but was '%v'", right)
	}

	return leftString + rightString, nil
}

func (e *FunctionCallExpression) evaluate(env Env) (any, error) {
	builtin := builtins[e.Token.Lexeme]
	return builtin.Func(env, e.Arguments)
}

func (e *LiteralExpression) evaluate(env Env) (any, error) {
	return e.Token.Literal, nil
}

func (e *VariableExpression) evaluate(env Env) (any, error) {
	identifier := e.Token.Lexeme
	val, found := env[identifier]
	if found {
		return val, nil
	}
	return nil, fmt.Errorf("undefined variable '%s'", identifier)
}

func isWeirdlyTrue(v any) bool {
	return v != 0
}

func boolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func intToBool(i int) bool {
	if i == 0 {
		return false
	}
	return true
}
