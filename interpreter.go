package main

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
)

var ErrExitFunction = errors.New("exit function")
var ErrExitLoop = errors.New("exit loop")
var ErrNextIteration = errors.New("next iteration")

type Env map[string]any

type ToiInstance struct {
	toiType     *TypeStatement
	fieldValues []any
}

func (instance *ToiInstance) print(out *bytes.Buffer) {
	out.WriteString(instance.toiType.Identifier.Lexeme)
	out.WriteRune('{')
	for i := range len(instance.fieldValues) {
		fieldName := instance.toiType.Fields[i].Lexeme
		fieldValue := instance.fieldValues[i]
		if i != 0 {
			out.WriteRune(',')
		}
		out.WriteString(fieldName)
		out.WriteRune('=')
		writeValue(fieldValue, out)
	}
	out.WriteRune('}')
}

const outerScope = "_outer"

// TODO: global state is bad; also this seems rather inefficient
type LineCol struct{ line, col int }

var currentInterpreterLineCol LineCol

// Statements

func (s *BlockStatement) execute(env Env) error {
	currentInterpreterLineCol = s.lineCol()
	for _, stmt := range s.Statements {
		if err := stmt.execute(env); err != nil {
			return err
		}
	}
	return nil
}

func (s *TypeStatement) execute(env Env) error {
	currentInterpreterLineCol = s.lineCol()

	env[getFuncEnvName(s.Identifier.Lexeme)] = s
	return nil
}

func (s *IfStatement) execute(env Env) error {
	currentInterpreterLineCol = s.lineCol()
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
	currentInterpreterLineCol = s.lineCol()
	for {
		v, err := s.Condition.evaluate(env)
		if err != nil {
			return err
		}
		if !isWeirdlyTrue(v) {
			break
		}

		if err := s.Body.execute(env); err != nil {
			if errors.Is(err, ErrExitLoop) {
				return nil
			} else if !errors.Is(err, ErrNextIteration) {
				return err
			}
		}

		if s.AfterBody != nil {
			if err := s.AfterBody.execute(env); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ExitFunctionStatement) execute(env Env) error {
	currentInterpreterLineCol = s.lineCol()
	return ErrExitFunction
}

func (s *ExitLoopStatement) execute(env Env) error {
	currentInterpreterLineCol = s.lineCol()
	return ErrExitLoop
}

func (s *NextIterationStatement) execute(env Env) error {
	currentInterpreterLineCol = s.lineCol()
	return ErrNextIteration
}

func (s *FunctionDeclarationStatement) execute(env Env) error {
	currentInterpreterLineCol = s.lineCol()

	env[getFuncEnvName(s.Identifier.Lexeme)] = s
	return nil
}

func (s *AssignmentStatement) execute(env Env) error {
	currentInterpreterLineCol = s.lineCol()
	v, err := s.Expression.evaluate(env)
	if err != nil {
		return err
	}
	identifier := s.Identifier.Lexeme
	env[identifier] = v
	return nil
}

func (s *ExpressionStatement) execute(env Env) error {
	currentInterpreterLineCol = s.lineCol()
	_, err := s.Expression.evaluate(env) /* Discard return value */
	return err
}

// Expressions

func (e *BinaryExpression) evaluate(env Env) (any, error) {
	currentInterpreterLineCol = e.lineCol()
	if e.Operator.Type == TokenOr {
		return e.evaluateOrOrAnd(env, isWeirdlyTrue)
	} else if e.Operator.Type == TokenAnd {
		return e.evaluateOrOrAnd(env, func(v any) bool { return !isWeirdlyTrue(v) })
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
	case TokenPercent:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return l % r })
	case TokenBAnd:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return l & r })
	case TokenBOr:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return l | r })
	case TokenXOr:
		return intBinaryOp(left, right, operator, func(l int, r int) int { return l ^ r })

	case TokenUnderscore:
		return stringConcat(left, right)

	case TokenEqualEqual:
		return boolToInt(isEqual(left, right)), nil
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

func (e *BinaryExpression) evaluateOrOrAnd(env Env, testFunc func(v any) bool) (any, error) {
	currentInterpreterLineCol = e.lineCol()
	left, err := e.Left.evaluate(env)
	if err != nil {
		return nil, err
	}

	leftInt, err := castToInt(left, "left", e.Operator.Lexeme)
	if err != nil {
		return nil, err
	}

	if testFunc(leftInt) {
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

func isEqual(left, right any) bool {
	leftArray, ok := left.(*[]any)
	if ok {
		rightArray, ok := right.(*[]any)
		if ok {
			return slices.Equal(*leftArray, *rightArray)
		}
	}

	return left == right
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

func (e *FieldAccessExpression) evaluate(env Env) (any, error) {
	left, err := e.Left.evaluate(env)
	if err != nil {
		return nil, err
	}
	instance, ok := left.(*ToiInstance)
	if !ok {
		return nil, fmt.Errorf("left-hand operand of '.' must be a type instance but was '%v'", left)
	}
	for i, field := range instance.toiType.Fields {
		if field.Lexeme == e.Identifier.Lexeme {
			return instance.fieldValues[i], nil
		}
	}
	return nil, fmt.Errorf("field '%v' not found on type '%v'", e.Identifier.Lexeme, instance.toiType.Identifier.Lexeme)
}

func (e *ContainerAccessExpression) evaluate(env Env) (any, error) {
	currentInterpreterLineCol = e.lineCol()
	get := builtins["get"]
	return get.Func(env, []Expression{e.Container, e.Access})
}

func (e *FunctionCallExpression) evaluate(env Env) (any, error) {
	currentInterpreterLineCol = e.lineCol()
	if e.Builtin {
		builtin := builtins[e.FunctionName]
		return builtin.Func(env, e.Arguments)
	}

	var envToFindFunc = env
	if outer, found := env[outerScope]; found {
		envToFindFunc = outer.(Env)
	}
	stmt := envToFindFunc[getFuncEnvName(e.FunctionName)]

	if fff, ok := stmt.(*FunctionDeclarationStatement); ok {
		funcStmt := fff
		functionEnv := make(Env)
		if outer, found := env[outerScope]; found {
			functionEnv[outerScope] = outer
		} else {
			functionEnv[outerScope] = env
		}

		if funcStmt.OutVariable != nil {
			functionEnv[funcStmt.OutVariable.Lexeme] = nil
		}
		var err error
		for i, param := range funcStmt.Parameters {
			functionEnv[param.Lexeme], err = e.Arguments[i].evaluate(env)
			if err != nil {
				return nil, err
			}
		}
		if err = funcStmt.Body.execute(functionEnv); err != nil {
			if !errors.Is(err, ErrExitFunction) {
				return nil, err
			}
		}
		if funcStmt.OutVariable != nil {
			return functionEnv[funcStmt.OutVariable.Lexeme], nil
		}
		return nil, nil
	} else {
		typeStmt := stmt.(*TypeStatement)
		if len(typeStmt.Fields) != len(e.Arguments) {
			panic(fmt.Sprintf("%d != %d", len(typeStmt.Fields), len(e.Arguments)))
		}

		fieldValues := make([]any, len(typeStmt.Fields))
		for i := range typeStmt.Fields {
			value, err := e.Arguments[i].evaluate(env)
			if err != nil {
				return nil, err
			}
			fieldValues[i] = value
		}

		return &ToiInstance{
			toiType:     typeStmt,
			fieldValues: fieldValues,
		}, nil
	}
}

func (e *LiteralExpression) evaluate(env Env) (any, error) {
	currentInterpreterLineCol = e.lineCol()
	return e.Token.Literal, nil
}

func (e *VariableExpression) evaluate(env Env) (any, error) {
	currentInterpreterLineCol = e.lineCol()
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

func getFuncEnvName(identifier string) string {
	return "_func_" + identifier
}
