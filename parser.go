package main

import "fmt"

type Statement func()
type Expression func() float64

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

func parseStatement(tokens []Token) (Statement, []Token, error) {
	if tokens[0].Type == TokenNewline {
		// Skip over empty lines
		return nil, tokens[1:], nil
	}

	if tokens[0].Type == TokenPrint {
		if len(tokens) < 2 {
			return nil, tokens[1:], fmt.Errorf("expected expression after 'print'")
		}

		expr, next, err := parseExpression(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		if len(next) != 0 && next[0].Type != TokenNewline {
			return nil, tokens[2:], fmt.Errorf("expected newline after statement but got %s ('%s')", tokens[2].Type, tokens[2].Lexeme)
		}

		if len(tokens) != 0 {
			// Consume newline
			next = next[1:]
		}
		return func() { fmt.Printf("%v\n", expr()) }, next, nil
	}

	return nil, tokens[1:], fmt.Errorf("unexpected token %s", tokens[0].Lexeme)
}

func parseExpression(tokens []Token) (Expression, []Token, error) {
	left, next, err := parseNumber(tokens)
	if err != nil {
		return nil, nil, err
	}
	tokens = next

	for len(tokens) != 0 && tokens[0].Type == TokenPlus {
		right, next, err := parseNumber(tokens[1:])
		if err != nil {
			return nil, nil, err
		}

		leftHand := left
		left = func() float64 {
			return leftHand() + right()
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

	return func() float64 { return tokens[0].Literal.(float64) }, tokens[1:], nil
}
