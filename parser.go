package main

import "fmt"

type Env map[string]any
type Statement func(env Env)
type Expression func(env Env) int

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

	if tokens[0].Type == TokenPrint {
		stmt, next, err = parsePrintStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else if tokens[0].Type == TokenWhile {
		stmt, next, err = parseWhileStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else if tokens[0].Type == TokenIdentifier {
		stmt, next, err = parseAssignmentStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else {
		return nil, tokens[1:], fmt.Errorf("unexpected token %s ('%s')", tokens[0].Type, tokens[0].Lexeme)
	}

	if len(next) != 0 && next[0].Type != TokenNewline && next[0].Type != TokenBraceClose {
		return nil, next, fmt.Errorf("expected newline after statement but got %s ('%s')", next[0].Type, next[0].Lexeme)
	}

	if len(next) != 0 && next[0].Type == TokenNewline {
		// Consume newline
		next = next[1:]
	}
	return stmt, next, nil
}

func parsePrintStatement(tokens []Token) (Statement, []Token, error) {
	if len(tokens) < 2 {
		return nil, tokens[1:], fmt.Errorf("expected expression after 'print'")
	}

	expr, next, err := parseExpression(tokens[1:])
	if err != nil {
		return nil, nil, err
	}

	return func(env Env) { fmt.Printf("%v\n", expr(env)) }, next, nil
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
		return nil, next, fmt.Errorf("expected '{' after while expression")
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
		return nil, next, fmt.Errorf("expected '}' after while statements")
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
		return nil, tokens[1:], fmt.Errorf("expected '=' and expression after identifier")
	} else if tokens[1].Type != TokenEquals {
		return nil, tokens[2:], fmt.Errorf("expected '=' after identifier")
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
	left, next, err := parsePlus(tokens)
	if err != nil {
		return nil, nil, err
	}
	tokens = next

	for len(tokens) != 0 && tokens[0].Type == TokenPlus {
		right, next, err := parsePlus(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		leftHand := left
		left = func(env Env) int {
			return leftHand(env) + right(env)
		}

		tokens = next
	}

	return left, tokens, nil
}

func parsePlus(tokens []Token) (Expression, []Token, error) {
	left, next, err := parseDivide(tokens)
	if err != nil {
		return nil, nil, err
	}
	tokens = next

	for len(tokens) != 0 && tokens[0].Type == TokenMinus {
		right, next, err := parseDivide(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		leftHand := left
		left = func(env Env) int {
			return leftHand(env) - right(env)
		}

		tokens = next
	}

	return left, tokens, nil
}

func parseDivide(tokens []Token) (Expression, []Token, error) {
	left, next, err := parsePrimary(tokens)
	if err != nil {
		return nil, nil, err
	}
	tokens = next

	for len(tokens) != 0 && tokens[0].Type == TokenSlash {
		right, next, err := parsePrimary(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		leftHand := left
		left = func(env Env) int {
			return leftHand(env) / right(env)
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
		return func(env Env) int {
			val, found := env[identifier]
			if found {
				return val.(int)
			}
			return 0
		}, tokens[1:], nil
	} else if token.Type == TokenInputNumber {
		if len(tokens) < 4 || tokens[1].Type != TokenParenOpen || (tokens[2].Type != TokenNumber && tokens[2].Type != TokenIdentifier) && tokens[3].Type != TokenBraceClose {
			return nil, tokens[1:], fmt.Errorf("expected a parenthesized number or identifier after `inputNumber`, but got %s ('%s')", token.Type, token.Lexeme)
		}

		var indexFunc func(Env) int
		if tokens[2].Type == TokenIdentifier {
			identifier := tokens[2].Lexeme
			indexFunc = func(env Env) int { return env[identifier].(int) }
		} else {
			index := tokens[2].Literal.(int)
			indexFunc = func(Env) int { return index }
		}
		return func(env Env) int { return env["_inputNumbers"].([]int)[indexFunc(env)] }, tokens[4:], nil
	}

	return nil, nil, fmt.Errorf("expected primary expression but got %s ('%s')", token.Type, token.Lexeme)
}
