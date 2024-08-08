package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 1 && len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Expected either:")
		fmt.Fprintln(os.Stderr, "\tNo arguments: provide script in stdin")
		fmt.Fprintln(os.Stderr, "\t1 argument:   script file (so stdin can be fed to the script)")
		os.Exit(1)
		return
	}

	var scriptData []byte
	var err error
	var stdin []byte
	if len(os.Args) == 1 {
		scriptData, err = io.ReadAll(os.Stdin)
	} else {
		scriptData, err = os.ReadFile(os.Args[1])
		ohno(err)
		stdin, err = io.ReadAll(os.Stdin)
		ohno(err)
	}
	ohno(err)

	tokens, errors := tokenize(string(scriptData))
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

	vars := make(map[string]any)

	var inputNumbersCache []int = nil
	populateInputNumbersCache := func() error {
		inputNumbersCache = make([]int, 0)
		for _, line := range strings.Split(string(stdin), "\n") {
			if line == "" {
				continue
			}
			n, err := strconv.Atoi(line)
			if err != nil {
				return err
			}
			inputNumbersCache = append(inputNumbersCache, n)
		}
		return nil
	}

	getInputNumbers := func() ([]int, error) {
		if inputNumbersCache == nil {
			if err := populateInputNumbersCache(); err != nil {
				return nil, err
			}
		}
		return inputNumbersCache, nil
	}

	vars["_getInputNumbers"] = getInputNumbers
	vars["_stdin"] = stdin
	for _, s := range statements {
		if err := s.execute(vars); err != nil {
			fmt.Fprintf(os.Stderr, "Execution error:\n\t%v\n", err)
			os.Exit(1)
			return
		}
	}
}

func ohno(err error) {
	if err != nil {
		panic(err)
	}
}
