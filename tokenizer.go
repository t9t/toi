package main

import (
	"fmt"
	"strings"
)

type TokenType string

const (
	TokenPrint   TokenType = "Print"
	TokenNewline TokenType = "Newline"
	TokenDigit   TokenType = "Digit"
)

type Token struct {
	Type   TokenType
	Lexeme string
}

func tokenize(input string) (tokens []Token, errors []error) {
	addToken := func(token Token) {
		tokens = append(tokens, token)
	}

	addError := func(err error) {
		errors = append(errors, err)
	}

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		c := runes[i]

		switch c {
		case ' ', '\t':
			break
		case '\n', '\r':
			addToken(Token{TokenNewline, string(c)})
		default:
			if c >= '0' && c <= '9' {
				addToken(Token{TokenDigit, string(c)})
			} else if c == 'p' {
				if strings.HasPrefix(string(runes[i:]), "print") {
					addToken(Token{TokenPrint, "print"})
					i += 4
				} else {
					addError(fmt.Errorf("expected 'rint' after 'p'"))
				}
			} else {
				addError(fmt.Errorf("unexpected character: %c", c))
			}
		}
	}

	return
}
