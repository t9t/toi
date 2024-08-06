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

	isAlpha := func(c rune) bool { return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
	isAlphaNum := func(c rune) bool { return isDigit(c) || isAlpha(c) }

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		c := runes[i]

		switch {
		case c == ' ' || c == '\t':
			break
		case c == '\n' || c == '\r':
			addToken(Token{TokenNewline, string(c)})
		case isDigit(c):
			token, err := tokenizeNumber(runes[i:])
			if err != nil {
				addError(err)
				break
			}

			addToken(token)
			i += len(token.Lexeme) - 1
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

func tokenizeNumber(runes []rune) (Token, error) {
	if runes[0] == '0' {
		return Token{}, fmt.Errorf("numbers may not start with 0")
	}

	i := 0
	for ; i < len(runes) && isDigit(runes[i]); i++ {
	}

	if len(runes[i:]) >= 2 && runes[i] == '.' && isDigit(runes[i+1]) {
		i += 1 // Consume the .
		for ; i < len(runes) && isDigit(runes[i]); i++ {
		}
	}

	return Token{TokenNumber, string(runes[0:i])}, nil
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}
