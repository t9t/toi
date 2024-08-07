package main

import (
	"slices"
)

func day01part1(input string) any {
	nums := numbers(input)
	slices.Sort(nums)

	target := 2020
	l, r := 0, len(nums)-1
	for l < r {
		left, right := nums[l], nums[r]
		sum := left + right
		if sum == target {
			return left * right
		} else if sum > target {
			r -= 1
		} else {
			l += 1
		}
	}
	panic("no answer found")
}

func day01part2(input string) any {
	nums := numbers(input)
	for i, left := range nums {
		for j := i + 1; j < len(nums); j++ {
			middle := nums[j]
			for k := j + 1; k < len(nums); k++ {
				right := nums[k]
				if left+middle+right == 2020 {
					return left * middle * right
				}
			}
		}
	}
	panic("no answer found")
}
