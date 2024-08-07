package main

import (
	"fmt"
	"strconv"
)

type TokenType string

const (
	TokenNewline    TokenType = "Newline"
	TokenIdentifier TokenType = "Identifier"
	TokenNumber     TokenType = "Number"
	TokenPlus       TokenType = "Plus"
	TokenSlash      TokenType = "Slash"
	TokenMinus      TokenType = "Minus"
	TokenEquals     TokenType = "Equals"

	TokenParenOpen  TokenType = "ParenOpen"
	TokenParenClose TokenType = "ParenClose"
	TokenBraceOpen  TokenType = "BraceOpen"
	TokenBraceClose TokenType = "BraceClose"

	TokenComma TokenType = "Comma"

	TokenIf    TokenType = "If"
	TokenWhile TokenType = "While"
)

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any
}

var singleCharTokens = map[rune]TokenType{
	'\n': TokenNewline,
	'\r': TokenNewline,

	'(': TokenParenOpen,
	')': TokenParenClose,
	'{': TokenBraceOpen,
	'}': TokenBraceClose,

	',': TokenComma,

	'+': TokenPlus,
	'/': TokenSlash,
	'-': TokenMinus,
	'=': TokenEquals,
}

var keywordTokens = map[string]TokenType{
	"if":    TokenIf,
	"while": TokenWhile,
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

		tokenType, found := singleCharTokens[c]
		if found {
			addToken(Token{tokenType, string(c), nil})
			continue
		}

		switch {
		case c == ' ' || c == '\t':
			break
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
			tokenType, found := keywordTokens[identifier]
			if found {
				addToken(Token{tokenType, identifier, nil})
			} else {
				addToken(Token{TokenIdentifier, identifier, nil})
			}
			i = j - 1
		default:
			addError(fmt.Errorf("unexpected character: %c", c))
		}
	}

	return
}

func tokenizeNumber(runes []rune) (Token, error) {
	i := 0
	for ; i < len(runes) && isDigit(runes[i]); i++ {
	}

	/* TODO: float support disabled for now lol
	if len(runes[i:]) >= 2 && runes[i] == '.' && isDigit(runes[i+1]) {
		i += 1 // Consume the .
		for ; i < len(runes) && isDigit(runes[i]); i++ {
		}
	}
	*/

	lexeme := string(runes[0:i])
	literal, err := strconv.Atoi(lexeme)
	if err != nil {
		// TODO: better errors for really big numbers
		panic(fmt.Sprintf("error converting '%s' to int: %v", lexeme, err))
	}

	if len(lexeme) > 1 && lexeme[0] == '0' && lexeme[1] != '.' {
		return Token{}, fmt.Errorf("numbers may not start with 0")
	}

	return Token{TokenNumber, lexeme, literal}, nil
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}
