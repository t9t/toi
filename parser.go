package main

import "fmt"

type Env map[string]any
type Statement func(Env)
type Expression func(Env) int
type BuiltinFunc func(Env, []Expression) int

type Builtin struct {
	Arity int
	Func  BuiltinFunc
}

var builtins = map[string]Builtin{
	"println": {1, func(env Env, e []Expression) int {
		v := e[0](env)
		fmt.Printf("%v\n", v)
		return v
	}},
	"inputNumber": {1, func(env Env, e []Expression) int { return env["_inputNumbers"].([]int)[e[0](env)] }},
}

func parse(tokens []Token) (statements []Statement, err error) {
	for len(tokens) > 0 {
		stmt, next, err := parseStatement(tokens)
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
		tokens = next
	}

	return
}

func parseStatement(tokens []Token) (stmt Statement, next []Token, err error) {
	if tokens[0].Type == TokenNewline {
		// Skip over empty lines
		return nil, tokens[1:], nil
	}

	if tokens[0].Type == TokenWhile {
		stmt, next, err = parseWhileStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else if len(tokens) >= 2 && tokens[0].Type == TokenIdentifier && tokens[1].Type == TokenEquals {
		stmt, next, err = parseAssignmentStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else {
		expr, nextTokens, err := parseExpression(tokens)
		if err != nil {
			return nil, nil, err
		}
		stmt = func(env Env) { expr(env) /* Discard return value */ }
		next = nextTokens
	}

	if len(next) != 0 && next[0].Type != TokenNewline && next[0].Type != TokenBraceClose {
		return nil, nil, fmt.Errorf("expected newline after statement but got %s ('%s')", next[0].Type, next[0].Lexeme)
	}

	if len(next) != 0 && next[0].Type == TokenNewline {
		// Consume newline
		next = next[1:]
	}
	return stmt, next, nil
}

func parseWhileStatement(tokens []Token) (Statement, []Token, error) {
	expr, next, err := parseExpression(tokens[1:])
	if err != nil {
		return nil, nil, err
	}

	// TODO: parse blocks
	if len(next) == 0 || next[0].Type != TokenBraceOpen {
		if len(next) != 0 {
			next = next[1:]
		}
		return nil, nil, fmt.Errorf("expected '{' after while expression")
	}

	next = next[1:]
	statements := make([]Statement, 0)
	for len(next) != 0 && next[0].Type != TokenBraceClose {
		stmt, next2, err := parseStatement(next)
		if err != nil {
			return nil, next2, err
		} else if stmt == nil {
			next = next2
			continue
		}
		statements = append(statements, stmt)
		next = next2
	}

	if len(next) == 0 || next[0].Type != TokenBraceClose {
		if len(next) != 0 {
			next = next[1:]
		}
		return nil, nil, fmt.Errorf("expected '}' after while statements")
	}

	return func(env Env) {
		for expr(env) > 0 {
			for _, stmt := range statements {
				stmt(env)
			}
		}
	}, next[1:], nil
}

func parseAssignmentStatement(tokens []Token) (Statement, []Token, error) {
	if len(tokens) < 3 {
		return nil, nil, fmt.Errorf("expected '=' and expression after identifier")
	} else if tokens[1].Type != TokenEquals {
		return nil, nil, fmt.Errorf("expected '=' after identifier")
	}

	expr, next, err := parseExpression(tokens[2:])
	if err != nil {
		return nil, nil, err
	}

	identifier := tokens[0].Lexeme
	return func(env Env) { env[identifier] = expr(env) }, next, nil
}

func parseExpression(tokens []Token) (Expression, []Token, error) {
	return parseMinus(tokens)
}

func parseMinus(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenMinus, parsePlus, func(l, r int) int { return l - r })
}

func parsePlus(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenPlus, parseDivide, func(l, r int) int { return l + r })
}

func parseDivide(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenSlash, parsePrimary, func(l, r int) int { return l / r })
}

func parseBinary(tokens []Token, tokenType TokenType, down func([]Token) (Expression, []Token, error), op func(int, int) int) (Expression, []Token, error) {
	left, next, err := down(tokens)
	if err != nil {
		return nil, nil, err
	}
	tokens = next

	for len(tokens) != 0 && tokens[0].Type == tokenType {
		right, next, err := down(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		leftHand := left
		left = func(env Env) int {
			return op(leftHand(env), right(env))
		}

		tokens = next
	}

	return left, tokens, nil
}

func parsePrimary(tokens []Token) (Expression, []Token, error) {
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("expected primary expression but reached end of data")
	}

	token := tokens[0]
	if token.Type == TokenNumber {
		value := tokens[0].Literal.(int)
		return func(Env) int { return value }, tokens[1:], nil
	} else if token.Type == TokenIdentifier {
		identifier := token.Lexeme

		if len(tokens) >= 2 && tokens[1].Type == TokenParenOpen {
			return parseFunctionCall(tokens)
		}

		// Variable access
		return func(env Env) int {
			val, found := env[identifier]
			if found {
				return val.(int)
			}
			return 0
		}, tokens[1:], nil
	}

	return nil, nil, fmt.Errorf("expected primary expression but got %s ('%s')", token.Type, token.Lexeme)
}

func parseFunctionCall(tokens []Token) (Expression, []Token, error) {
	identifier := tokens[0].Lexeme

	builtin, found := builtins[identifier]
	if !found {
		return nil, nil, fmt.Errorf("no such builtin function '%s'", identifier)
	}

	tokens = tokens[2:] // Consume '('

	arguments := make([]Expression, 0)
	for len(tokens) > 0 {
		expr, next, err := parseExpression(tokens)
		if err != nil {
			return nil, nil, err
		}

		arguments = append(arguments, expr)
		tokens = next
		if len(tokens) > 0 {
			if tokens[0].Type == TokenParenClose {
				tokens = tokens[1:]
				break
			}
			if tokens[0].Type == TokenComma {
				tokens = tokens[1:]
				continue
			}
			return nil, nil, fmt.Errorf("expected ')' or ',' but got %s ('%s')", tokens[0].Type, tokens[0].Lexeme)
		}
		return nil, nil, fmt.Errorf("expected ')' or ',' but got end of input")
	}

	if len(arguments) != builtin.Arity {
		return nil, nil, fmt.Errorf("expected %d arguments but got %d for function '%s'", builtin.Arity, len(arguments), identifier)
	}

	return func(env Env) int { return builtin.Func(env, arguments) }, tokens, nil
}
