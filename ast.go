package main

type Statement interface {
	execute(env Env) error
	compile(compiler *Compiler) error

	lineCol() LineCol
}

type Expression interface {
	evaluate(env Env) (any, error)
	compile(compiler *Compiler) error

	lineCol() LineCol
}

type BlockStatement struct {
	Token      Token
	Statements []Statement
}

func (s *BlockStatement) lineCol() LineCol {
	return s.Token.LineCol()
}

type IfStatement struct {
	Token     Token
	Condition Expression
	Then      Statement
	Otherwise *Statement
}

func (s *IfStatement) lineCol() LineCol {
	return s.Token.LineCol()
}

type WhileStatement struct {
	Token     Token
	Condition Expression
	Body      Statement
	AfterBody Statement // For 'for' loops
}

func (s *WhileStatement) lineCol() LineCol {
	return s.Token.LineCol()
}

type ExitFunctionStatement struct {
	Token Token
}

func (s *ExitFunctionStatement) lineCol() LineCol {
	return s.Token.LineCol()
}

type ExitLoopStatement struct {
	Token Token
}

func (s *ExitLoopStatement) lineCol() LineCol {
	return s.Token.LineCol()
}

type NextIterationStatement struct {
	Token Token
}

func (s *NextIterationStatement) lineCol() LineCol {
	return s.Token.LineCol()
}

type FunctionDeclarationStatement struct {
	Identifier  Token
	Parameters  []Token
	OutVariable *Token
	Body        Statement
}

func (s *FunctionDeclarationStatement) lineCol() LineCol {
	return s.Identifier.LineCol()
}

type AssignmentStatement struct {
	Identifier Token
	Expression Expression
}

func (s *AssignmentStatement) lineCol() LineCol {
	return s.Identifier.LineCol()
}

type ExpressionStatement struct {
	Token      Token
	Expression Expression
}

func (s *ExpressionStatement) lineCol() LineCol {
	return s.Token.LineCol()
}

type BinaryExpression struct {
	Left     Expression
	Operator Token
	Right    Expression
}

func (e *BinaryExpression) lineCol() LineCol {
	return e.Operator.LineCol()
}

type ContainerAccessExpression struct {
	Token     Token
	Container Expression
	Access    Expression
}

func (e *ContainerAccessExpression) lineCol() LineCol {
	return e.Token.LineCol()
}

type FunctionCallExpression struct {
	Token        Token
	Builtin      bool
	FunctionName string
	Arguments    []Expression
}

func (e *FunctionCallExpression) lineCol() LineCol {
	return e.Token.LineCol()
}

type LiteralExpression struct {
	Token Token
}

func (e *LiteralExpression) lineCol() LineCol {
	return e.Token.LineCol()
}

type VariableExpression struct {
	Token Token
}

func (e *VariableExpression) lineCol() LineCol {
	return e.Token.LineCol()
}
