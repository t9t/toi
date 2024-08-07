package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type AocFunc func(string) any

var funcs = map[string]AocFunc{
	"1-1": day01part1,
	"1-2": day01part2,
}

func main() {
	if len(os.Args) != 3 {
		panic("invalid arguments; need day + part")
	}

	day, part := os.Args[1], os.Args[2]
	fmt.Printf("Running day %s; part %s\n", day, part)

	input, err := os.ReadFile("../aoc/input/2020/" + day + ".txt")
	rip(err)

	f, found := funcs[day+"-"+part]
	if !found {
		fatal("no func found for day %s part %s", day, part)
	}
	fmt.Println(f(string(input)))
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
