package main

import (
	"fmt"
	"io"
	"os"
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

func main() {
	data, err := io.ReadAll(os.Stdin)
	ohno(err)

	tokens := make([]Token, 0)

	for i := 0; i < len(data); i++ {
		b := data[i]
		if b == ' ' || b == '\t' {
			continue
		}

		if b == '\r' || b == '\n' {
			tokens = append(tokens, Token{TokenNewline, string(b)})
		} else if b >= '0' && b <= '9' {
			tokens = append(tokens, Token{TokenDigit, string(b)})
		} else if b == 'p' {
			if strings.HasPrefix(string(data[i:]), "print") {
				tokens = append(tokens, Token{TokenPrint, "print"})
				i += 4
			} else {
				panic("invalid input, expected 'rint' after 'p'")
			}
		} else {
			panic("invalid input, unexpected byte '" + string(b) + "'")
		}
	}

	fmt.Printf("Tokens: %v\n", tokens)

	statements := make([]func(), 0)

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		fmt.Printf("Token %d %+v\n", i, t)
		if t.Type == TokenNewline {
			continue
		}

		if t.Type != TokenPrint {
			panic("syntax error, expected 'print'")
		}

		if len(tokens[i:]) < 3 {
			panic("syntax error, expected digit and newline after 'print'")
		}

		next := tokens[i+1]
		nextNext := tokens[i+2]

		if next.Type != TokenDigit || nextNext.Type != TokenNewline {
			panic("syntax error, expected digit and newline after 'print'")
		}

		i += 2

		statements = append(statements, func() { fmt.Printf("%s\n", next.Lexeme) })
	}

	fmt.Printf("Statements: %v\n", statements)

	for _, s := range statements {
		s()
	}
}

func ohno(err error) {
	if err != nil {
		panic(err)
	}
}
