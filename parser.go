package main

import "fmt"

type Statement func()

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
			return nil, tokens[1:], fmt.Errorf("expected value after 'print'")
		}

		next := tokens[1]
		if next.Type != TokenInt && next.Type != TokenFloat {
			return nil, tokens[1:], fmt.Errorf("expected number after 'print' but got %s ('%s')", next.Type, next.Lexeme)
		}

		if len(tokens) >= 3 && tokens[2].Type != TokenNewline {
			return nil, tokens[2:], fmt.Errorf("expected newline after statement but got %s ('%s')", tokens[2].Type, tokens[2].Lexeme)
		}

		offset := 2
		if len(tokens) >= 3 {
			offset = 3
		}
		return func() { fmt.Printf("%v\n", next.Literal) }, tokens[offset:], nil
	}

	return nil, tokens[1:], fmt.Errorf("unexpected token %s", tokens[0].Lexeme)
}
