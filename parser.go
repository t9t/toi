package main

import (
	"fmt"
)

type Env map[string]any

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

	if tokens[0].Type == TokenIf {
		stmt, next, err = parseIfStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else if tokens[0].Type == TokenWhile {
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
		stmt = &ExpressionStatement{Expression: expr}
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

func parseBlock(tokens []Token, typ string) (Statement, []Token, error) {
	next := tokens

	// TODO: parse blocks
	if len(next) == 0 || next[0].Type != TokenBraceOpen {
		if len(next) != 0 {
			next = next[1:]
		}
		return nil, nil, fmt.Errorf("expected '{' after %s expression", typ)
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
		return nil, nil, fmt.Errorf("expected '}' after %s statements", typ)
	}

	return &BlockStatement{Statements: statements}, next[1:], nil
}

func parseIfStatement(tokens []Token) (Statement, []Token, error) {
	expr, next, err := parseExpression(tokens[1:])
	if err != nil {
		return nil, nil, err
	}

	block, next, err := parseBlock(next, "if")
	if err != nil {
		return nil, nil, err
	}

	return &IfStatement{Condition: expr, Then: block}, next, nil
}

func parseWhileStatement(tokens []Token) (Statement, []Token, error) {
	expr, next, err := parseExpression(tokens[1:])
	if err != nil {
		return nil, nil, err
	}

	block, next, err := parseBlock(next, "while")
	if err != nil {
		return nil, nil, err
	}

	return &WhileStatement{Condition: expr, Body: block}, next, nil
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

	return &AssignmentStatement{Identifier: tokens[0], Expression: expr}, next, nil
}

func parseExpression(tokens []Token) (Expression, []Token, error) {
	return parseEqualEqual(tokens)
}

func parseEqualEqual(tokens []Token) (Expression, []Token, error) {
	return parseBinary2(tokens, TokenEqualEqual, parseNotEqual)
}

func parseNotEqual(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenNotEqual, parseGreaterThan)
}

func parseGreaterThan(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenGreaterThan, parseGreaterEqual)
}

func parseGreaterEqual(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenGreaterEqual, parseLessThan)
}

func parseLessThan(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenLessThan, parseLessEqual)
}

func parseLessEqual(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenLessEqual, parseMinus)
}

func parseMinus(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenMinus, parsePlus)
}

func parsePlus(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenPlus, parseDivide)
}

func parseDivide(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenSlash, parseMultiply)
}

func parseMultiply(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenAsterisk, parseUnderscore)
}

func parseUnderscore(tokens []Token) (Expression, []Token, error) {
	left, next, err := parsePrimary(tokens)
	if err != nil {
		return nil, nil, err
	}
	tokens = next

	for len(tokens) != 0 && tokens[0].Type == TokenUnderscore {
		right, next, err := parsePrimary(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		left = &BinaryExpression{Left: left, Operator: tokens[0], Right: right}
		tokens = next
	}

	return left, tokens, nil
}

func parseBinary(tokens []Token, tokenType TokenType, down func([]Token) (Expression, []Token, error)) (Expression, []Token, error) {
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

		left = &BinaryExpression{Left: left, Operator: tokens[0], Right: right}
		tokens = next
	}

	return left, tokens, nil
}

func parseBinary2(tokens []Token, tokenType TokenType, down func([]Token) (Expression, []Token, error)) (Expression, []Token, error) {
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

		left = &BinaryExpression{Left: left, Operator: tokens[0], Right: right}
		tokens = next
	}

	return left, tokens, nil
}

func parsePrimary(tokens []Token) (Expression, []Token, error) {
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("expected primary expression but reached end of data")
	}

	token := tokens[0]
	if token.Type == TokenString || token.Type == TokenNumber {
		return &LiteralExpression{Token: token}, tokens[1:], nil
	} else if token.Type == TokenIdentifier {
		if len(tokens) >= 2 && tokens[1].Type == TokenParenOpen {
			return parseFunctionCall(tokens)
		}

		// Variable access
		return &VariableExpression{Token: token}, tokens[1:], nil
	}

	return nil, nil, fmt.Errorf("expected primary expression but got %s ('%s')", token.Type, token.Lexeme)
}

func parseFunctionCall(tokens []Token) (Expression, []Token, error) {
	callToken := tokens[0]
	identifier := callToken.Lexeme

	builtin, found := builtins[identifier]
	if !found {
		return nil, nil, fmt.Errorf("no such builtin function '%s'", identifier)
	}

	tokens = tokens[2:] // Consume identifier and '('

	arguments := make([]Expression, 0)
	for len(tokens) > 0 {
		// TODO: remove duplication
		if tokens[0].Type == TokenParenClose {
			tokens = tokens[1:]
			break
		}

		expr, next, err := parseExpression(tokens)
		if err != nil {
			return nil, nil, err
		}

		arguments = append(arguments, expr)
		tokens = next
		if len(tokens) > 0 {
			if tokens[0].Type == TokenComma {
				tokens = tokens[1:]
			} else if tokens[0].Type != TokenParenClose {
				return nil, nil, fmt.Errorf("expected ')' or ',' but got %s ('%s')", tokens[0].Type, tokens[0].Lexeme)
			}
		} else {
			return nil, nil, fmt.Errorf("expected ')' or ',' but got end of input")
		}
	}

	if len(arguments) != builtin.Arity && builtin.Arity != ArityVariadic {
		return nil, nil, fmt.Errorf("expected %d arguments but got %d for function '%s'", builtin.Arity, len(arguments), identifier)
	}

	return &FunctionCallExpression{Token: callToken, Arguments: arguments}, tokens, nil
}
