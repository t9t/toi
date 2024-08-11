package main

type Statement interface {
	execute(env Env) error
	compile() ([]byte, error)
}

type Expression interface {
	evaluate(env Env) (any, error)
	compile() ([]byte, error)
}

type BlockStatement struct {
	Statements []Statement
}

type IfStatement struct {
	Condition Expression
	Then      Statement
}

type WhileStatement struct {
	Condition Expression
	Body      Statement
}

type AssignmentStatement struct {
	Identifier Token
	Expression Expression
}

type ExpressionStatement struct {
	Expression Expression
}

type BinaryExpression struct {
	Left     Expression
	Operator Token
	Right    Expression
}

type FunctionCallExpression struct {
	Token     Token
	Arguments []Expression
}

type LiteralExpression struct {
	Token Token
}

type VariableExpression struct {
	Token Token
}
