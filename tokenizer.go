package main

import (
	"fmt"
	"strconv"
	"strings"
)

type TokenType string

const (
	TokenNewline         TokenType = "Newline"
	TokenIdentifier      TokenType = "Identifier"
	TokenNumber          TokenType = "Number"
	TokenString          TokenType = "String"
	TokenInternalLiteral TokenType = "InternalLiteral"

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
	TokenPercent    TokenType = "Percent"
	TokenUnderscore TokenType = "Underscore"

	TokenAmpersand TokenType = "Ampersand"
	TokenPipe      TokenType = "Pipe"

	TokenOr   TokenType = "Or"
	TokenAnd  TokenType = "And"
	TokenBOr  TokenType = "BOr"
	TokenXOr  TokenType = "XOr"
	TokenBAnd TokenType = "BAnd"

	TokenEquals TokenType = "Equals"

	TokenParenOpen    TokenType = "ParenOpen"
	TokenParenClose   TokenType = "ParenClose"
	TokenBraceOpen    TokenType = "BraceOpen"
	TokenBraceClose   TokenType = "BraceClose"
	TokenBracketOpen  TokenType = "BracketOpen"
	TokenBracketClose TokenType = "BracketClose"

	TokenComma TokenType = "Comma"

	TokenIf        TokenType = "If"
	TokenOtherwise TokenType = "Otherwise"
	TokenWhile     TokenType = "While"
	TokenFor       TokenType = "For"

	TokenExit      TokenType = "Exit"
	TokenFunction  TokenType = "Function"
	TokenLoop      TokenType = "Loop"
	TokenNext      TokenType = "Next"
	TokenIteration TokenType = "Iteration"

	TokenFullStop TokenType = "FullStop"
)

type Token struct {
	Type           TokenType
	Lexeme         string
	Literal        any
	Pos, Line, Col int
}

func (t *Token) LineCol() LineCol {
	return LineCol{t.Line, t.Col}
}

var singleCharTokens = map[rune]TokenType{
	'\n': TokenNewline,
	'\r': TokenNewline,

	'(': TokenParenOpen,
	')': TokenParenClose,
	'{': TokenBraceOpen,
	'}': TokenBraceClose,
	'[': TokenBracketOpen,
	']': TokenBracketClose,

	',': TokenComma,

	'+': TokenPlus,
	'_': TokenUnderscore,
	'-': TokenMinus,
	'*': TokenAsterisk,
	'%': TokenPercent,

	'&': TokenAmpersand,
	'|': TokenPipe,

	'.': TokenFullStop,
}

var keywordTokens = map[string]TokenType{
	"if":        TokenIf,
	"otherwise": TokenOtherwise,
	"while":     TokenWhile,
	"for":       TokenFor,
	"exit":      TokenExit,
	"function":  TokenFunction,
	"loop":      TokenLoop,
	"next":      TokenNext,
	"iteration": TokenIteration,
	"or":        TokenOr,
	"and":       TokenAnd,
	"bor":       TokenBOr,
	"xor":       TokenXOr,
	"band":      TokenBAnd,
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

	// col is set to 0 because we increment them right after reading
	line, col := 1, 0

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		col += 1

		tokenType, found := singleCharTokens[c]
		if found {
			addToken(Token{tokenType, string(c), nil, i, line, col})
			if c == '\n' || c == '\r' {
				line += 1
				col = 0
			}
			continue
		}

		switch {
		case c == ' ' || c == '\t':
			break
		case c == '/':
			if i != len(runes)-1 && runes[i+1] == '/' {
				// Comment
				j := i + 1
				// we don't increment col, because we don't expect any tokens (and thus errors) after //
				if runes[j+1] == '\r' || runes[j+1] == '\n' {
					// Commenting out the newline
					i = j + 1
					col = 0
					line += 1
				} else {
					// Discard until end of line; but keep the newline
					for ; j < len(runes) && (runes[j] != '\r' && runes[j] != '\n'); j++ {
					}
					i = j - 1
				}
			} else {
				addToken(Token{TokenSlash, "/", nil, i, line, col})
			}
		case c == '=':
			if i != len(runes)-1 && runes[i+1] == '=' {
				i += 1
				col += 1
				addToken(Token{TokenEqualEqual, "==", nil, i, line, col})
			} else {
				addToken(Token{TokenEquals, "=", nil, i, line, col})
			}
		case c == '>':
			if i != len(runes)-1 && runes[i+1] == '=' {
				i += 1
				col += 1
				addToken(Token{TokenGreaterEqual, ">=", nil, i, line, col})
			} else {
				addToken(Token{TokenGreaterThan, ">", nil, i, line, col})
			}
		case c == '<':
			if i != len(runes)-1 && runes[i+1] == '=' {
				i += 1
				col += 1
				addToken(Token{TokenLessEqual, "<=", nil, i, line, col})
			} else if i != len(runes)-1 && runes[i+1] == '>' {
				i += 1
				col += 1
				addToken(Token{TokenNotEqual, "<>", nil, i, line, col})
			} else {
				addToken(Token{TokenLessThan, "<", nil, i, line, col})
			}
		case c == '"':
			token, err := tokenizeString(runes[i+1:], i, line, col)
			if err != nil {
				addError(err)
				break
			}
			addToken(token)
			i += len(token.Lexeme) + 1
			col += len(token.Lexeme) + 1
		case isDigit(c):
			token, err := tokenizeNumber(runes[i:], i, line, col)
			if err != nil {
				addError(err)
				break
			}

			addToken(token)
			i += len(token.Lexeme) - 1
			col += len(token.Lexeme) - 1
		case isAlpha(c):
			j := i + 1
			for ; j < len(runes) && isAlphaNum(runes[j]); j++ {
			}
			identifier := string(runes[i:j])
			tokenType, found := keywordTokens[identifier]
			if found {
				addToken(Token{tokenType, identifier, nil, i, line, col})
			} else {
				addToken(Token{TokenIdentifier, identifier, nil, i, line, col})
			}
			newI := j - 1
			col += (newI - i)
			i = newI
		default:
			addError(fmt.Errorf("unexpected character: %c (line %d, col %d)", c, line, col))
		}
	}

	return
}

func tokenizeString(runes []rune, pos, line, col int) (Token, error) {
	i := 0
	for i < len(runes) {
		if runes[i] == '"' {
			break
		}
		if i < len(runes)-4 && runes[i] == '$' && runes[i+1] == '{' && runes[i+2] == '"' && runes[i+3] == '}' {
			i += 3
		}
		i += 1
	}

	if i == len(runes) || runes[i] != '"' {
		return Token{}, fmt.Errorf("unterminated string at %d:%d", line, col)
	}

	lexeme := string(runes[0:i])
	literal := strings.ReplaceAll(lexeme, "${\"}", "\"")

	return Token{TokenString, lexeme, literal, pos, line, col}, nil
}

func tokenizeNumber(runes []rune, pos, line, col int) (Token, error) {
	i := 0
	for ; i < len(runes) && (isDigit(runes[i]) || runes[i] == '\''); i++ {
	}

	/* TODO: float support disabled for now lol
	if len(runes[i:]) >= 2 && runes[i] == '.' && isDigit(runes[i+1]) {
		i += 1 // Consume the .
		for ; i < len(runes) && isDigit(runes[i]); i++ {
		}
	}
	*/

	rawLexeme := string(runes[0:i])
	fixedLexeme := strings.ReplaceAll(rawLexeme, "'", "")
	literal, err := strconv.Atoi(fixedLexeme)
	if err != nil {
		// TODO: better errors for really big numbers
		return Token{}, fmt.Errorf("error converting '%s' to int: %v", fixedLexeme, err)
	}

	if len(fixedLexeme) > 1 && fixedLexeme[0] == '0' && fixedLexeme[1] != '.' {
		return Token{}, fmt.Errorf("numbers may not start with 0")
	}

	return Token{TokenNumber, rawLexeme, literal, pos, line, col}, nil
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}
