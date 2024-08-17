package main

import (
	"slices"
	"strconv"
	"strings"
)

func day10part2(input string) any {
	inputNumbers := []int{0}
	for _, line := range strings.Split(strings.TrimSpace(input), "\n") {
		n, err := strconv.Atoi(line)
		rip(err)
		inputNumbers = append(inputNumbers, n)
	}
	slices.Sort(inputNumbers)
	inputNumbers = append(inputNumbers, inputNumbers[len(inputNumbers)-1]+3)

	cache := make(map[int]int)
	return determine(inputNumbers, cache)
}

func determine(numbers []int, cache map[int]int) int {
	if len(numbers) == 2 {
		// There's always just one way to go from the penultimate to the last one
		return 1
	}
	first := numbers[0]
	if count, found := cache[first]; found {
		return count
	}

	total := 0
	if len(numbers) >= 4 && numbers[3]-first <= 3 {
		// yes: [0, x, x, 3, ...]; no: [0, x, x, 5, ...]
		total += determine(numbers[3:], cache)

	}
	if len(numbers) >= 3 && numbers[2]-first <= 3 {
		// yes: [0, x, 3, ...]; no: [0, x, 5, ...]
		total += determine(numbers[2:], cache)

	}
	if len(numbers) >= 2 && numbers[1]-first <= 3 {
		// yes: [0, 3, ...]; no: [0, 5, ...]
		total += determine(numbers[1:], cache)

	}

	cache[first] = total

	return total
}
