package main

import (
	"strconv"
	"strings"
)

func day01part1(input string) any {
	sum := 0
	for _, s := range strings.Split(input, "\n") {
		if s == "" {
			continue
		}
		n, err := strconv.Atoi(s)
		rip(err)
		sum += n/3 - 2
	}
	return sum
}
