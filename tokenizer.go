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
	TokenString     TokenType = "String"

	TokenEqualEqual   TokenType = "EqualEqual"
	TokenNotEqual     TokenType = "NotEqual"
	TokenGreaterThan  TokenType = "GreaterThan"
	TokenGreaterEqual TokenType = "GreaterEqual"
	TokenLessThan     TokenType = "LessThan"
	TokenLessEqual    TokenType = "LessEqual"

	TokenPlus       TokenType = "Plus"
	TokenMinus      TokenType = "Minus"
	TokenAsterisk   TokenType = "Asterisk"
	TokenSlash      TokenType = "Slash"
	TokenUnderscore TokenType = "Underscore"

	TokenAmpersand TokenType = "Ampersand"

	TokenEquals TokenType = "Equals"

	TokenParenOpen  TokenType = "ParenOpen"
	TokenParenClose TokenType = "ParenClose"
	TokenBraceOpen  TokenType = "BraceOpen"
	TokenBraceClose TokenType = "BraceClose"

	TokenComma TokenType = "Comma"

	TokenIf        TokenType = "If"
	TokenOtherwise TokenType = "Otherwise"
	TokenWhile     TokenType = "While"
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
	'_': TokenUnderscore,
	'-': TokenMinus,
	'*': TokenAsterisk,

	'&': TokenAmpersand,
}

var keywordTokens = map[string]TokenType{
	"if":        TokenIf,
	"otherwise": TokenOtherwise,
	"while":     TokenWhile,
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
		case c == '/':
			if i != len(runes)-1 && runes[i+1] == '/' {
				// Comment
				j := i + 1
				if runes[j+1] == '\r' || runes[j+1] == '\n' {
					// Commenting out the newline
					i = j + 1
				} else {
					// Discard until end of line; but keep the newline
					for ; j < len(runes) && (runes[j] != '\r' && runes[j] != '\n'); j++ {
					}
					i = j - 1
				}
			} else {
				addToken(Token{TokenSlash, "/", nil})
			}
		case c == '=':
			if i != len(runes)-1 && runes[i+1] == '=' {
				i += 1
				addToken(Token{TokenEqualEqual, "==", nil})
			} else {
				addToken(Token{TokenEquals, "=", nil})
			}
		case c == '>':
			if i != len(runes)-1 && runes[i+1] == '=' {
				i += 1
				addToken(Token{TokenGreaterEqual, ">=", nil})
			} else {
				addToken(Token{TokenGreaterThan, ">", nil})
			}
		case c == '<':
			if i != len(runes)-1 && runes[i+1] == '=' {
				i += 1
				addToken(Token{TokenLessEqual, "<=", nil})
			} else if i != len(runes)-1 && runes[i+1] == '>' {
				i += 1
				addToken(Token{TokenNotEqual, "<>", nil})
			} else {
				addToken(Token{TokenLessThan, "<", nil})
			}
		case c == '"':
			token := tokenizeString(runes[i+1:])
			addToken(token)
			i += len(token.Lexeme) + 1
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

func tokenizeString(runes []rune) Token {
	i := 0
	for ; i < len(runes) && runes[i] != '"'; i++ {
	}

	lexeme := string(runes[0:i])
	literal := lexeme

	return Token{TokenString, lexeme, literal}
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
		return Token{}, fmt.Errorf("error converting '%s' to int: %v", lexeme, err)
	}

	if len(lexeme) > 1 && lexeme[0] == '0' && lexeme[1] != '.' {
		return Token{}, fmt.Errorf("numbers may not start with 0")
	}

	return Token{TokenNumber, lexeme, literal}, nil
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}
