package main

import (
	"fmt"
)

type TokenType string

const (
	TokenPrint   TokenType = "Print"
	TokenNewline TokenType = "Newline"
	TokenNumber  TokenType = "Number"
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

	isNum := func(c rune) bool { return c >= '0' && c <= '9' }
	isAlpha := func(c rune) bool { return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
	isAlphaNum := func(c rune) bool { return isNum(c) || isAlpha(c) }

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		fmt.Printf("Token %d: #%c#\n", i, c)

		switch {
		case c == ' ' || c == '\t':
			break
		case c == '\n' || c == '\r':
			addToken(Token{TokenNewline, string(c)})
		case isNum(c):
			j := i + 1
			for ; j < len(runes) && isNum(runes[j]); j++ {
			}
			addToken(Token{TokenNumber, string(runes[i:j])})
			i = j - 1
		case isAlpha(c):
			j := i + 1
			for ; j < len(runes) && isAlphaNum(runes[j]); j++ {
			}
			identifier := string(runes[i:j])
			if identifier == "print" {
				addToken(Token{TokenPrint, identifier})
			} else {
				addError(fmt.Errorf("unsupported identifier '%s'", identifier))
			}
			i = j - 1
		default:
			addError(fmt.Errorf("unexpected character: %c", c))
		}
	}

	return
}
