package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type AocFunc func(string) any

var funcs = map[string]AocFunc{
	"1-1": day01part1,
	"1-2": day01part2,
	"7-2": day07part2,
}

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		panic("invalid arguments; need day + part and optionall input file")
	}

	day, part := os.Args[1], os.Args[2]
	fmt.Printf("Running day %s; part %s\n", day, part)

	inputFile := "../aoc/input/2020/" + day + ".txt"
	if len(os.Args) == 4 {
		inputFile = os.Args[3]
	}

	input, err := os.ReadFile(inputFile)
	rip(err)

	f, found := funcs[day+"-"+part]
	if !found {
		fatal("no func found for day %s part %s", day, part)
	}
	inputString := string(input)
	start := time.Now()
	out := f(inputString)
	took := time.Since(start)

	fmt.Printf("Took: %v\n%v\n", took, out)
}

func rip(err error) {
	if err != nil {
		panic(err)
	}
}

func fatal(format string, a ...any) {
	panic(fmt.Sprintf(format, a...))
}

func lines(s string) []string {
	return strings.Split(strings.TrimSpace(s), "\n")
}

func numbers(s string) (nums []int) {
	for _, line := range lines(s) {
		n, err := strconv.Atoi(line)
		rip(err)
		nums = append(nums, n)
	}
	return
}
