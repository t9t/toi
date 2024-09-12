package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

// TODO: definitely not globals
var toiStdout *bytes.Buffer
var toiStdin string

func main() {
	args := os.Args[1:] // strip command
	outFile := ""
	if len(args) != 0 && args[0] == "-o" {
		if len(args) == 1 {
			printUsageAndExit()
		}
		outFile = args[1]
		args = args[2:]
	}

	if len(args) > 1 {
		printUsageAndExit()
	}

	var stdout string
	var err error
	stdin, err := io.ReadAll(os.Stdin)
	ohno(err)

	var scriptName string
	if len(args) == 0 {
		scriptName = "(stdin)"
		stdout, err = runScript(stdin, outFile, "")
	} else if len(args) == 1 {
		scriptName = args[0]
		stdout, err = runScriptFile(scriptName, outFile, string(stdin))
	}

	fmt.Print(stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing script '%s': %v\n", scriptName, err)
		os.Exit(1)
	}
	return
}

func runScript(scriptData []byte, outFile string, stdin string) (string, error) {
	tokens, errors := tokenize(string(scriptData))
	if len(errors) != 0 {
		fmt.Fprintf(os.Stderr, "Got %d errors:\n", len(errors))
		for i, err := range errors {
			fmt.Fprintf(os.Stderr, "  %d: %v\n", i, err)
		}
		return "", fmt.Errorf("tokenization error")
	}

	parser := &Parser{tokens: tokens, declaredFunctions: make(map[string]int), declaredTypes: make(map[string]struct{})}
	scriptStatement, err := parser.parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	// TODO: better state management instead of globals
	toiStdin = stdin
	toiStdout = &bytes.Buffer{}

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

	compiler := &Compiler{functions: make(map[string]VmFunction), declaredTypes: make(map[string]VmType)}
	err = scriptStatement.compile(compiler)
	if err != nil {
		return "", fmt.Errorf("Compilation error: %w", err)
	}
	ops := compiler.bytes
	//decompile(compiler.constants, ops)

	if outFile != "" {
		err = dump(outFile, ops, compiler.constants, compiler.variables, compiler.functions)
		ohno(err)
	}

	start = time.Now()
	err = execute(ops, compiler.constants, compiler.variables, compiler.functions, compiler.declaredTypes)
	if err != nil {
		fmt.Printf("%s\n", toiStdout.String())
		return "", fmt.Errorf("VM execution error: %w", err)
	}
	// TODO: kinda annoying that it also counts initialization time
	fmt.Printf("VM run time: %v\n", time.Since(start))

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

func runScriptFile(filepath string, outFile string, stdin string) (string, error) {
	var scriptData []byte
	var err error
	scriptData, err = os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	return runScript(scriptData, outFile, stdin)
}

func ohno(err error) {
	if err != nil {
		panic(err)
	}
}

func printUsageAndExit() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-o outfile] [script file]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "    -o outfile:    write the produced bytcode to the <outfile>\n")
	fmt.Fprintf(os.Stderr, "    script file:   run the script file; if not provided, provide the script in stdin\n")
	os.Exit(1)
	return
}
