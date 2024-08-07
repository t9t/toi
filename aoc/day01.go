package main

import (
	"strconv"
	"strings"
)

func day01part1(input string) any {
	nums := make([]int, 0)
	for _, s := range strings.Split(input, "\n") {
		if s == "" {
			continue
		}
		n, err := strconv.Atoi(s)
		rip(err)
		nums = append(nums, n)
	}

	for i, left := range nums {
		for j := i + 1; j < len(nums); j++ {
			right := nums[j]
			if left+right == 2020 {
				return left * right
			}
		}
	}
	panic("no answer found")
}

func day01part2(input string) any {
	sum := 0
	payloads := strings.Split(input, "\n")
	for i := 0; i < len(payloads); i++ {
		s := payloads[i]
		if s == "" {
			continue
		}
		n, err := strconv.Atoi(s)
		rip(err)
		fuel := n/3 - 2
		if fuel > 0 {
			sum += fuel
			payloads = append(payloads, strconv.Itoa(fuel))
		}
	}
	return sum
}
