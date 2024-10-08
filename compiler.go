package main

import (
	"fmt"
)

// TODO: use a bytebuffer instead of slices for efficiency; although slices are nice and easy to patch jumps

type LoopState struct {
	exitLoops      []int
	nextIterations []int
}

type Compiler struct {
	constants []any
	bytes     []byte
	variables []string // index = id; value = name

	loopStates    []*LoopState
	functions     map[string]VmFunction
	exitFunctions []int
	declaredTypes map[string]VmType
}

func (c *Compiler) writeByte(b byte) {
	c.bytes = append(c.bytes, b)
}

func (c *Compiler) setByte(index int, b byte) {
	c.bytes[index] = b
}

func (c *Compiler) writeBytes(bytes ...byte) {
	c.bytes = append(c.bytes, bytes...)
}

func (c *Compiler) currentLoopState() *LoopState {
	return c.loopStates[len(c.loopStates)-1]
}

func (c *Compiler) pushLoopState() {
	c.loopStates = append(c.loopStates, &LoopState{})
}

func (c *Compiler) popLoopState() {
	// leaves some memory filled, but we're OK with that
	c.loopStates = c.loopStates[0 : len(c.loopStates)-1]
}

func (c *Compiler) addExitLoop(index int) {
	loopState := c.currentLoopState()
	loopState.exitLoops = append(loopState.exitLoops, index)
}

func (c *Compiler) addNextIteration(index int) {
	loopState := c.currentLoopState()
	loopState.nextIterations = append(loopState.nextIterations, index)
}

func (c *Compiler) len() int {
	return len(c.bytes)
}

// Statements

func (s *BlockStatement) compile(compiler *Compiler) error {
	for _, stmt := range s.Statements {
		if err := stmt.compile(compiler); err != nil {
			return err
		}
	}
	return nil
}

func (s *TypeStatement) compile(compiler *Compiler) error {
	fields := make([]string, len(s.Fields))
	fieldMap := make(map[string]int, len(s.Fields))
	for i, fieldToken := range s.Fields {
		fields[i] = fieldToken.Lexeme
		fieldMap[fieldToken.Lexeme] = i
	}

	vmType := VmType{
		Name:     s.Identifier.Lexeme,
		Fields:   fields,
		FieldMap: fieldMap,
	}
	compiler.declaredTypes[vmType.Name] = vmType

	return nil
}

func (s *IfStatement) compile(compiler *Compiler) error {
	if err := s.Condition.compile(compiler); err != nil {
		return err
	}

	thenJumpIndex := compiler.len()
	compiler.writeBytes(OpJumpIfFalse, InvalidOp, InvalidOp)

	if err := s.Then.compile(compiler); err != nil {
		return err
	}

	thenJumpTo := compiler.len()
	jumpOverOtherwiseIndex := compiler.len()
	if s.Otherwise != nil {
		compiler.writeBytes(OpJumpForward, InvalidOp, InvalidOp)
		thenJumpTo = compiler.len()

		if err := (*s.Otherwise).compile(compiler); err != nil {
			return err
		}
	}

	thenJumpAmount := thenJumpTo - thenJumpIndex - 3 // I don't know why it must be 3
	b1, b2, err := encodeJumpAmount(thenJumpAmount)
	if err != nil {
		// TODO: add token/line/col to error
		return err
	}
	compiler.setByte(thenJumpIndex+1, b1)
	compiler.setByte(thenJumpIndex+2, b2)

	if s.Otherwise != nil {
		otherwiseJumpAmount := compiler.len() - jumpOverOtherwiseIndex - 3 // why 3 though?
		b1, b2, err := encodeJumpAmount(otherwiseJumpAmount)
		if err != nil {
			// TODO: add token/line/col to error
			return err
		}
		compiler.setByte(jumpOverOtherwiseIndex+1, b1)
		compiler.setByte(jumpOverOtherwiseIndex+2, b2)
	}

	return nil
}

func (s *WhileStatement) compile(compiler *Compiler) error {
	compiler.pushLoopState()

	conditionIndex := compiler.len()
	if err := s.Condition.compile(compiler); err != nil {
		return err
	}
	conditionFalseJumpIndex := compiler.len()
	compiler.writeBytes(OpJumpIfFalse, InvalidOp, InvalidOp)

	if err := s.Body.compile(compiler); err != nil {
		return err
	}

	afterBodyIndex := compiler.len()
	if s.AfterBody != nil { // for loop incrementor
		if err := s.AfterBody.compile(compiler); err != nil {
			return err
		}
	}

	jumpBackAmount := compiler.len() - conditionIndex + 3 // I don't get why it's 3
	b1, b2, err := encodeJumpAmount(jumpBackAmount)
	if err != nil {
		// TODO: error reporting with token/line/col
		return err
	}
	compiler.writeBytes(OpJumpBack, b1, b2)

	// Patch jump over loop
	jumpForwardAmount := compiler.len() - conditionFalseJumpIndex - 3 // 3!?
	b1, b2, err = encodeJumpAmount(jumpForwardAmount)
	if err != nil {
		// TODO: error reporting with token/line/col
		return err
	}
	compiler.setByte(conditionFalseJumpIndex+1, b1)
	compiler.setByte(conditionFalseJumpIndex+2, b2)

	var nextIterationJumpIndex int
	if s.AfterBody != nil {
		nextIterationJumpIndex = afterBodyIndex
	} else {
		nextIterationJumpIndex = conditionIndex
	}

	loopState := compiler.currentLoopState()
	for _, index := range loopState.nextIterations {
		var jumpAmount int
		if s.AfterBody != nil {
			// For loop; jump to incrementor
			compiler.setByte(index, OpJumpForward)
			jumpAmount = nextIterationJumpIndex - index - 3 // I don't understand why it must be 3
		} else {
			// While loop; just jump to expression
			compiler.setByte(index, OpJumpBack)
			jumpAmount = index - nextIterationJumpIndex + 3 // I don't understand why it must be 3
		}

		b1, b2, err := encodeJumpAmount(jumpAmount)
		if err != nil {
			// TODO: error reporting with token/line/col
			return err
		}
		compiler.setByte(index+1, b1)
		compiler.setByte(index+2, b2)
	}

	endOfLoopIndex := compiler.len()
	for _, index := range loopState.exitLoops {
		jumpAmount := endOfLoopIndex - index - 3 // I don't understand why it must be 3
		b1, b2, err := encodeJumpAmount(jumpAmount)
		if err != nil {
			// TODO: error reporting with token/line/col
			return err
		}
		compiler.setByte(index+1, b1)
		compiler.setByte(index+2, b2)
	}

	compiler.popLoopState()
	return nil
}

func (s *ExitFunctionStatement) compile(compiler *Compiler) error {
	compiler.exitFunctions = append(compiler.exitFunctions, compiler.len())
	compiler.writeBytes(OpJumpForward, InvalidOp, InvalidOp)
	return nil
}

func (s *ExitLoopStatement) compile(compiler *Compiler) error {
	compiler.addExitLoop(compiler.len())
	compiler.writeBytes(OpJumpForward, InvalidOp, InvalidOp)
	return nil
}

func (s *NextIterationStatement) compile(compiler *Compiler) error {
	compiler.addNextIteration(compiler.len())
	// Jump type set in parseWhileStatement (back for while; forward for for)
	compiler.writeBytes(InvalidOp, InvalidOp, InvalidOp)
	return nil
}

func encodeJumpAmount(amount int) (byte, byte, error) {
	if amount > MaxBlockSize {
		// TODO: keep tokens for error reporting
		return 0, 0, fmt.Errorf("jump of %d exceeds maximum of %d operations", amount, MaxBlockSize)
	}
	return byte(amount / 256), byte(amount % 256), nil
}

func (s *FunctionDeclarationStatement) compile(compiler *Compiler) error {
	hasOutVar := s.OutVariable != nil
	outVarIdentifier := ""
	if hasOutVar {
		outVarIdentifier = s.OutVariable.Lexeme
	}

	functionVariables := make([]string, len(s.Parameters))
	for i, param := range s.Parameters {
		functionVariables[i] = param.Lexeme
	}
	if hasOutVar {
		functionVariables = append(functionVariables, outVarIdentifier)
	}

	functionCompiler := &Compiler{constants: compiler.constants, functions: compiler.functions, variables: functionVariables}
	if err := s.Body.compile(functionCompiler); err != nil {
		return err
	}

	endOfLoopIndex := functionCompiler.len()
	for _, index := range functionCompiler.exitFunctions {
		jumpAmount := endOfLoopIndex - index - 3 // I don't understand why it must be 3
		b1, b2, err := encodeJumpAmount(jumpAmount)
		if err != nil {
			// TODO: error reporting with token/line/col
			return err
		}
		functionCompiler.setByte(index+1, b1)
		functionCompiler.setByte(index+2, b2)
	}

	ops := functionCompiler.bytes
	compiler.constants = functionCompiler.constants
	params := make([]string, len(s.Parameters))
	for i, param := range s.Parameters {
		params[i] = param.Lexeme
	}
	compiler.functions[s.Identifier.Lexeme] = VmFunction{params: params, ops: ops, variableDefinitions: functionCompiler.variables, hasOutVar: hasOutVar}

	return nil
}

func (s *AssignmentStatement) compile(compiler *Compiler) error {
	index, err := compiler.registerVariable(s.Identifier.Lexeme)
	if err != nil {
		return err
	}

	if err := s.Expression.compile(compiler); err != nil {
		return err
	}

	compiler.writeBytes(OpSetVariable, index)
	return nil
}

func (s *FieldAssignmentStatement) compile(compiler *Compiler) error {
	if err := s.Left.compile(compiler); err != nil {
		return err
	} else if err := s.Expression.compile(compiler); err != nil {
		return err
	}

	index, err := compiler.ensureConstant(s.Identifier.Lexeme)
	if err != nil {
		return err
	}

	compiler.writeBytes(OpSetField, index)
	return nil
}

func (s *ExpressionStatement) compile(compiler *Compiler) error {
	/* Discard return value afterwards using pop */
	if err := s.Expression.compile(compiler); err != nil {
		return err
	}
	compiler.writeByte(OpPop)
	return nil
}

// Expressions

func (e *BinaryExpression) compile(compiler *Compiler) error {
	if e.Operator.Type == TokenOr {
		return e.compileOrOrAnd(compiler, true)
	} else if e.Operator.Type == TokenAnd {
		return e.compileOrOrAnd(compiler, false)
	}

	if err := e.Left.compile(compiler); err != nil {
		return err
	} else if err := e.Right.compile(compiler); err != nil {
		return err
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
	case TokenPercent:
		binaryOp = OpBinaryRemainder
	case TokenBAnd:
		binaryOp = OpBinaryBinaryAnd
	case TokenBOr:
		binaryOp = OpBinaryBinaryOr
	case TokenXOr:
		binaryOp = OpBinaryBinaryXor

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
		return fmt.Errorf("unsupported binary operator %v ('%v')", e.Operator.Type, e.Operator.Lexeme)
	}

	compiler.writeBytes(OpBinary, binaryOp)
	if appendNot {
		compiler.writeByte(OpNot)
	}
	return nil
}

func (e *BinaryExpression) compileOrOrAnd(compiler *Compiler, withNot bool) error {
	if err := e.Left.compile(compiler); err != nil {
		return err
	}

	compiler.writeByte(OpDuplicate)
	if withNot {
		compiler.writeByte(OpNot)
	}

	jumpIndex := compiler.len()
	compiler.writeBytes(OpJumpIfFalse, InvalidOp, InvalidOp, OpPop)

	if err := e.Right.compile(compiler); err != nil {
		return err
	}

	jumpAmount := compiler.len() - jumpIndex - 3 // 2x jump offset + 1x pop
	b1, b2, err := encodeJumpAmount(jumpAmount)
	if err != nil {
		// TODO: add token/line/col to error
		return err
	}

	compiler.setByte(jumpIndex+1, b1)
	compiler.setByte(jumpIndex+2, b2)
	return nil
}

func (e *FieldAccessExpression) compile(compiler *Compiler) error {
	if err := e.Left.compile(compiler); err != nil {
		return err
	}

	index, err := compiler.ensureConstant(e.Identifier.Lexeme)
	if err != nil {
		return err
	}

	compiler.writeBytes(OpFieldAccess, index)
	return nil
}

func (e *ContainerAccessExpression) compile(compiler *Compiler) error {
	f := &FunctionCallExpression{
		Token:        e.Token,
		Builtin:      true,
		FunctionName: "get",
		Arguments:    []Expression{e.Container, e.Access},
	}
	return f.compile(compiler)
}

func (e *FunctionCallExpression) compile(compiler *Compiler) error {
	if len(e.Arguments) > 50 {
		return fmt.Errorf("functions don't support more than 50 arguments (was %d for '%v')", len(e.Arguments), e.FunctionName)
	}

	for _, arg := range e.Arguments {
		if err := arg.compile(compiler); err != nil {
			return err
		}
	}

	index, err := compiler.ensureConstant(e.FunctionName)
	if err != nil {
		return err
	}

	builtin, found := builtins[e.FunctionName]
	if found && builtin.Arity == ArityVariadic {
		compiler.writeBytes(OpCallVariadicFunction, index, byte(len(e.Arguments)))
		return nil
	}

	op := OpCallFunction
	if e.Builtin {
		op = OpCallBuiltin
	} else if e.Constructor {
		op = OpInstantiate
	}
	compiler.writeBytes(op, index)
	return nil
}

func (e *LiteralExpression) compile(compiler *Compiler) error {
	if i, ok := e.Token.Literal.(int); ok && i <= 0xFF {
		compiler.writeBytes(OpInlineNumber, byte(i))
		return nil
	}

	index, err := compiler.ensureConstant(e.Token.Literal)
	if err != nil {
		return err
	}
	compiler.writeBytes(OpLoadConstant, index)
	return nil
}

func (e *VariableExpression) compile(compiler *Compiler) error {
	identifier := e.Token.Lexeme
	index, found := compiler.findVariableIndex(identifier)
	if !found {
		tok := e.Token
		return fmt.Errorf("variable '%v' used before set at %d:%d", identifier, tok.Line, tok.Col)
	}

	compiler.writeBytes(OpReadVariable, index)

	return nil
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

func (c *Compiler) registerVariable(name string) (byte, error) {
	for i, v := range c.variables {
		if v == name {
			return byte(i), nil
		}
	}

	if len(c.constants) == MaxConstants {
		return 0, fmt.Errorf("cannot add variable '%v' because the maximum of %d was reached", name, MaxConstants)
	}

	c.variables = append(c.variables, name)
	return byte(len(c.variables) - 1), nil
}

func (c *Compiler) findVariableIndex(name string) (byte, bool) {
	for i, v := range c.variables {
		if v == name {
			return byte(i), true
		}
	}

	return 0, false
}
