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
	left, err := e.Left.evaluate(env)
	if err != nil {
		return nil, err
	}

	right, err := e.Right.evaluate(env)
	if err != nil {
		return nil, err
	}

	token := e.Operator

	switch e.Operator.Type {
	case TokenPlus:
		return intBinaryOp(left, right, token, func(l int, r int) int { return l + r })
	case TokenMinus:
		return intBinaryOp(left, right, token, func(l int, r int) int { return l - r })
	case TokenAsterisk:
		return intBinaryOp(left, right, token, func(l int, r int) int { return l * r })
	case TokenSlash:
		return intBinaryOp(left, right, token, func(l int, r int) int { return l / r })

	case TokenUnderscore:
		return stringConcat(left, right, token)

	case TokenEqualEqual:
		return boolToInt(left == right), nil
	case TokenNotEqual:
		return boolToInt(left != right), nil
	case TokenGreaterThan:
		return intBinaryOp(left, right, token, func(l int, r int) int { return boolToInt(l > r) })
	case TokenGreaterEqual:
		return intBinaryOp(left, right, token, func(l int, r int) int { return boolToInt(l >= r) })
	case TokenLessThan:
		return intBinaryOp(left, right, token, func(l int, r int) int { return boolToInt(l < r) })
	case TokenLessEqual:
		return intBinaryOp(left, right, token, func(l int, r int) int { return boolToInt(l <= r) })
	}

	return nil, nil
}

func intBinaryOp(left, right any, operator Token, op func(int, int) int) (any, error) {
	leftInt, ok := left.(int)
	if !ok {
		return nil, fmt.Errorf("left-hand operand of '%s' should be an int but was '%v'", operator.Lexeme, left)
	}

	rightInt, ok := right.(int)
	if !ok {
		return nil, fmt.Errorf("right-hand operand of '%s' should be an int but was '%v'", operator.Lexeme, right)
	}

	return op(leftInt, rightInt), nil
}

func stringConcat(left, right any, operator Token) (any, error) {
	leftString, ok := left.(string)
	if !ok {
		return nil, fmt.Errorf("left-hand operand of '%s' should be a string but was '%v'", operator.Lexeme, left)
	}

	rightString, ok := right.(string)
	if !ok {
		return nil, fmt.Errorf("right-hand operand of '%s' should be a string but was '%v'", operator.Lexeme, right)
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
