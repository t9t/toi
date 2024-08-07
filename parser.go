package main

import "fmt"

type Statement func()
type Expression func() int

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
	} else if tokens[0].Type == TokenIdentifier {
		stmt, next, err = parseAssignmentStatement(tokens)
		if err != nil {
			return nil, next, err
		}
	} else {
		return nil, tokens[1:], fmt.Errorf("unexpected token %s", tokens[0].Lexeme)
	}

	if len(next) != 0 && next[0].Type != TokenNewline {
		return nil, next, fmt.Errorf("expected newline after statement but got %s ('%s')", next[0].Type, next[0].Lexeme)
	}

	if len(next) != 0 {
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

	return func() { fmt.Printf("%v\n", expr()) }, next, nil
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

	varName := tokens[0].Lexeme

	return func() { fmt.Printf("(assignment %v to %v)\n", expr(), varName) }, next, nil
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
		left = func() int {
			return leftHand() + right()
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
		left = func() int {
			return leftHand() - right()
		}

		tokens = next
	}

	return left, tokens, nil
}

func parseDivide(tokens []Token) (Expression, []Token, error) {
	left, next, err := parseNumber(tokens)
	if err != nil {
		return nil, nil, err
	}
	tokens = next

	for len(tokens) != 0 && tokens[0].Type == TokenSlash {
		right, next, err := parseNumber(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		leftHand := left
		left = func() int {
			return leftHand() / right()
		}

		tokens = next
	}

	return left, tokens, nil
}

func parseNumber(tokens []Token) (Expression, []Token, error) {
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("expected number but reached end of data")
	}

	token := tokens[0]
	if token.Type != TokenNumber {
		return nil, nil, fmt.Errorf("expected number but got %s ('%s')", token.Type, token.Lexeme)
	}

	return func() int { return tokens[0].Literal.(int) }, tokens[1:], nil
}
