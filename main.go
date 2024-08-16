package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

// TODO: definitely not globals
var toiStdout bytes.Buffer
var toiStdin string

func main() {
	if len(os.Args) != 1 && len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Expected either:")
		fmt.Fprintln(os.Stderr, "\tNo arguments: provide script in stdin")
		fmt.Fprintln(os.Stderr, "\t1 argument:   script file (so stdin can be fed to the script)")
		os.Exit(1)
		return
	}

	var stdout string
	var err error
	stdin, err := io.ReadAll(os.Stdin)
	ohno(err)

	var scriptName string
	if len(os.Args) == 1 {
		scriptName = "(stdin)"
		stdout, err = runScript(stdin, "")
	} else if len(os.Args) == 2 {
		scriptName = os.Args[1]
		stdout, err = runScriptFile(scriptName, string(stdin))
	}

	fmt.Print(stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing script '%s': %v\n", scriptName, err)
		os.Exit(1)
	}
	return
}

func runScript(scriptData []byte, stdin string) (string, error) {
	tokens, errors := tokenize(string(scriptData))
	if len(errors) != 0 {
		fmt.Fprintf(os.Stderr, "Got %d errors:\n", len(errors))
		for i, err := range errors {
			fmt.Fprintf(os.Stderr, "  %d: %v\n", i, err)
		}
		return "", fmt.Errorf("tokenization error")
	}

	parser := &Parser{tokens: tokens}
	scriptStatement, err := parser.parse(tokens)
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	// TODO: better state management instead of globals
	toiStdin = stdin
	toiStdout = bytes.Buffer{}

	vars := make(map[string]any)

	start := time.Now()
	if err := scriptStatement.execute(vars); err != nil {
		toiStdout.WriteTo(os.Stdout)
		fmt.Fprintf(os.Stderr, "Execution error at %d:%d:\n\t%v\n", currentInterpreterLineCol.line, currentInterpreterLineCol.col, err)
		return "", fmt.Errorf("execution error")
	}

	fmt.Printf("Tree interpreter run time: %v\n", time.Since(start))

	treeOutput := toiStdout.String()

	toiStdout.Reset()

	compiler := &Compiler{constants: make([]any, 0)}
	err = scriptStatement.compile(compiler)
	if err != nil {
		return "", fmt.Errorf("Compilation error: %w", err)
	}
	ops := compiler.bytes
	//decompile(compiler.constants, ops)
	err = execute(compiler.constants, ops)
	if err != nil {
		return "", fmt.Errorf("VM execution error: %w", err)
	}

	vmOutput := toiStdout.String()

	if vmOutput != treeOutput {
		fmt.Fprintln(os.Stderr, "Different output from VM than tree interpreter:")
		fmt.Fprintln(os.Stderr, "===== VM: =====")
		fmt.Fprintln(os.Stderr, vmOutput)
		fmt.Fprintln(os.Stderr, "===== Tree: =====")
		fmt.Fprintln(os.Stderr, treeOutput)
		panic("output mismatch")
	}

	return treeOutput, nil
}

func runScriptFile(filepath string, stdin string) (string, error) {
	var scriptData []byte
	var err error
	scriptData, err = os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	return runScript(scriptData, stdin)
}

func ohno(err error) {
	if err != nil {
		panic(err)
	}
}
