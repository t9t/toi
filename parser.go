package main

import (
	"fmt"
)

type Env map[string]any

// TODO: global state is yuck, don't do it
var whileBodyCount = 0

func parse(tokens []Token) (Statement, error) {
	statements := make([]Statement, 0)
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

	return &BlockStatement{Statements: statements}, nil
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
	} else if tokens[0].Type == TokenExit {
		stmt, next, err = parseExitLoopStatement(tokens)
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
		tok := next[0]
		return nil, nil, fmt.Errorf("expected newline after statement but got %s ('%s') at %d:%d", tok.Type, tok.Lexeme, tok.Line, tok.Col)
	}

	if len(next) != 0 && next[0].Type == TokenNewline {
		// Consume newline
		next = next[1:]
	}
	return stmt, next, nil
}

func parseBlock(tokens []Token, typ string) (Statement, []Token, error) {
	next := tokens

	if len(next) == 0 || next[0].Type != TokenBraceOpen {
		tok := next[0]
		return nil, nil, fmt.Errorf("expected '{' after %s at %d:%d", typ, tok.Line, tok.Col)
	}

	token := next[0]

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
		tok := next[0]
		return nil, nil, fmt.Errorf("expected '}' after %s statements at %d:%d", typ, tok.Line, tok.Col)
	}

	return &BlockStatement{Token: token, Statements: statements}, next[1:], nil
}

func parseIfStatement(tokens []Token) (Statement, []Token, error) {
	token := tokens[0]

	expr, next, err := parseExpression(tokens[1:])
	if err != nil {
		return nil, nil, err
	}

	block, next, err := parseBlock(next, "if expression")
	if err != nil {
		return nil, nil, err
	}

	var otherwiseBlock *Statement
	if len(next) > 0 && next[0].Type == TokenOtherwise {
		next = next[1:] // Consume "otherwise"
		otherwise, otherwiseNext, err := parseBlock(next, "otherwise")
		if err != nil {
			return nil, nil, err
		}
		next = otherwiseNext
		otherwiseBlock = &otherwise
	}

	return &IfStatement{Token: token, Condition: expr, Then: block, Otherwise: otherwiseBlock}, next, nil
}

func parseWhileStatement(tokens []Token) (Statement, []Token, error) {
	token := tokens[0]

	expr, next, err := parseExpression(tokens[1:])
	if err != nil {
		return nil, nil, err
	}

	whileBodyCount += 1
	block, next, err := parseBlock(next, "while expression")
	if err != nil {
		return nil, nil, err
	}
	whileBodyCount -= 1

	return &WhileStatement{Token: token, Condition: expr, Body: block}, next, nil
}

func parseExitLoopStatement(tokens []Token) (Statement, []Token, error) {
	token := tokens[0]

	if len(tokens) == 1 || tokens[1].Type != TokenLoop {
		tok := tokens[1]
		return nil, nil, fmt.Errorf("expected 'loop' after 'exit' at %d:%d", tok.Line, tok.Col)
	}

	if whileBodyCount == 0 {
		tok := tokens[0]
		return nil, nil, fmt.Errorf("can only use 'exit loop' in 'while' body at %d:%d", tok.Line, tok.Col)
	}

	return &ExitLoopStatement{Token: token}, tokens[2:], nil
}

func parseAssignmentStatement(tokens []Token) (Statement, []Token, error) {
	if len(tokens) < 3 {
		tok := tokens[0]
		return nil, nil, fmt.Errorf("expected '=' and expression after identifier at %d:%d", tok.Line, tok.Col)
	} else if tokens[1].Type != TokenEquals {
		tok := tokens[1]
		return nil, nil, fmt.Errorf("expected '=' after identifier at %d:%d", tok.Line, tok.Col)
	}

	expr, next, err := parseExpression(tokens[2:])
	if err != nil {
		return nil, nil, err
	}

	return &AssignmentStatement{Identifier: tokens[0], Expression: expr}, next, nil
}

func parseExpression(tokens []Token) (Expression, []Token, error) {
	return parseContainerAccess(tokens)
}

func parseContainerAccess(tokens []Token) (Expression, []Token, error) {
	isContainerAccess := false
	startToken := tokens[0]
	if tokens[0].Type == TokenBracketOpen {
		isContainerAccess = true
		tokens = tokens[1:]
	}

	expr, next, err := parseLogicalOr(tokens)
	if err != nil {
		return nil, nil, err
	} else if !isContainerAccess {
		return expr, next, nil
	}

	if next[0].Type != TokenBracketClose {
		tok := next[0]
		return nil, nil, fmt.Errorf("expected ']' after '[' and expression but got '%v' at %d:%d", tok.Type, tok.Line, tok.Col)
	}
	next = next[1:]

	indexExpr, more, err := parseExpression(next)
	if err != nil {
		return nil, nil, err
	}

	arguments := []Expression{expr, indexExpr}
	return &FunctionCallExpression{Token: startToken, FunctionName: "get", Arguments: arguments}, more, nil
}

func parseLogicalOr(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenPipe, parseLogicalAnd)
}

func parseLogicalAnd(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenAmpersand, parseEqualEqual)
}

func parseEqualEqual(tokens []Token) (Expression, []Token, error) {
	return parseBinary(tokens, TokenEqualEqual, parseNotEqual)
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
	return parseBinary(tokens, TokenUnderscore, parsePrimary)
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

	return nil, nil, fmt.Errorf("expected primary expression but got %s ('%s') at %d:%d", token.Type, token.Lexeme, token.Line, token.Col)
}

func parseFunctionCall(tokens []Token) (Expression, []Token, error) {
	callToken := tokens[0]
	identifier := callToken.Lexeme

	builtin, found := builtins[identifier]
	if !found {
		tok := callToken
		return nil, nil, fmt.Errorf("no such builtin function '%s' at %d:%d", identifier, tok.Line, tok.Col)
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
				tok := tokens[0]
				return nil, nil, fmt.Errorf("expected ')' or ',' but got %s ('%s') at %d:%d", tok.Type, tok.Lexeme, tok.Line, tok.Col)
			}
		} else {
			return nil, nil, fmt.Errorf("expected ')' or ',' but got end of input")
		}
	}

	if len(arguments) != builtin.Arity && builtin.Arity != ArityVariadic {
		tok := tokens[0]
		return nil, nil, fmt.Errorf("expected %d arguments but got %d for function '%s' at %d:%d", builtin.Arity, len(arguments), identifier, tok.Line, tok.Col)
	}

	return &FunctionCallExpression{Token: callToken, FunctionName: callToken.Lexeme, Arguments: arguments}, tokens, nil
}
