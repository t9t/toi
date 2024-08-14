package main

import (
	"fmt"
	"strconv"
	"strings"
)

type BagCount struct {
	Color string
	Count int
}

func day07part2(input string) any {
	lines := strings.Split(strings.TrimSpace(input), "\n")
	bags := make(map[string][]BagCount)

	for _, line := range lines {
		parts := strings.Split(line, " bags contain ")
		outer := parts[0]
		counts := make([]BagCount, 0)
		if strings.Contains(line, "no other bags") {
			bags[outer] = counts
			continue
		}
		parts = strings.Split(parts[1][:len(parts[1])-1], ", ")
		for _, part := range parts {
			bagParts := strings.Split(part, " ")
			num, _ := strconv.Atoi(bagParts[0])

			color := fmt.Sprintf("%s %s", bagParts[1], bagParts[2])
			counts = append(counts, BagCount{Color: color, Count: num})
		}
		bags[outer] = counts
	}

	search := []BagCount{{"shiny gold", 1}}
	total := 0
	for len(search) != 0 {
		item := search[0]
		search = search[1:]

		bagCounts := bags[item.Color]
		for _, bagCount := range bagCounts {
			search = append(search, BagCount{bagCount.Color, bagCount.Count * item.Count})
			total += bagCount.Count * item.Count
		}
	}
	return total
}

func countBags(m map[string][]BagCount, count int, color string) int {
	bagCounts := m[color]
	total := count
	for _, bagCount := range bagCounts {
		total += countBags(m, count*bagCount.Count, bagCount.Color)
	}
	return total
}
