package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) != 1 && len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Expected either:")
		fmt.Fprintln(os.Stderr, "\tNo arguments: provide script in stdin")
		fmt.Fprintln(os.Stderr, "\t1 argument:   script file (so stdin can be fed to the script)")
		os.Exit(1)
		return
	}

	var data []byte
	var err error
	if len(os.Args) == 1 {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(os.Args[1])
	}
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
