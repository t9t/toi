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
	if len(errors) != 0 {
		fmt.Fprintf(os.Stderr, "Got %d errors:\n", len(errors))
		for i, err := range errors {
			fmt.Fprintf(os.Stderr, "  %d: %v\n", i, err)
		}
		os.Exit(1)
		return
	}

	fmt.Printf("%d tokens: %v\n", len(tokens), tokens)

	statements, err := parse(tokens)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Printf("%d statements: %v\n", len(statements), statements)

	vars := make(map[string]int)
	for _, s := range statements {
		s(vars)
	}
}

func ohno(err error) {
	if err != nil {
		panic(err)
	}
}
