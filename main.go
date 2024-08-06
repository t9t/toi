package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	data, err := io.ReadAll(os.Stdin)
	ohno(err)

	tokens, errors := tokenize(string(data))
	if errors != nil {
		fmt.Printf("Got %d errors:\n", len(errors))
		for i, err := range errors {
			fmt.Printf("  %d: %v\n", i, err)
		}
		os.Exit(1)
		return
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

		if next.Type != TokenNumber || nextNext.Type != TokenNewline {
			panic("syntax error, expected number and newline after 'print'")
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
