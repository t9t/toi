package main

import (
	"fmt"
	"math"
	"strings"
)

type Coords struct{ x, y, z, w int }
type Grid map[Coords]struct{}

var deltas []Coords

func (g Grid) setActive(coords Coords) {
	g[coords] = struct{}{}
}

func (g Grid) getNeighbors(coords Coords) []Coords {
	neighbors := make([]Coords, len(deltas))
	for i, delta := range deltas {
		neighbors[i] = Coords{coords.x + delta.x, coords.y + delta.y, coords.z + delta.z, coords.w + delta.w}
	}
	return neighbors
}

func (g Grid) activeNeighborCount(coords Coords) int {
	count := 0
	for _, delta := range deltas {
		otherCoords := Coords{coords.x + delta.x, coords.y + delta.y, coords.z + delta.z, coords.w + delta.w}
		if _, found := g[otherCoords]; found {
			count += 1
		}
	}
	return count
}

func (g Grid) draw(heading string) {
	minX, minY, minZ, minW := math.MaxInt, math.MaxInt, math.MaxInt, math.MaxInt
	maxX, maxY, maxZ, maxW := math.MinInt, math.MinInt, math.MinInt, math.MinInt
	for coords, _ := range g {
		minX, minY, minZ, minW = min(minX, coords.x), min(minY, coords.y), min(minZ, coords.z), min(minW, coords.w)
		maxX, maxY, maxZ, maxW = max(maxX, coords.x), max(maxY, coords.y), max(maxZ, coords.z), max(maxW, coords.w)
	}

	fmt.Println()
	fmt.Println(heading)
	fmt.Println()
	for w := minW; w <= maxW; w += 1 {
		for z := minZ; z <= maxZ; z += 1 {
			fmt.Printf("z=%d, w=%d\n", z, w)
			for y := minY; y <= maxY; y += 1 {
				for x := minX; x <= maxX; x += 1 {
					c := '.'
					if _, found := g[Coords{x, y, z, w}]; found {
						c = '#'
					}
					fmt.Printf("%c", c)
				}
				fmt.Println()
			}
			fmt.Println()
		}
	}
}

func (g Grid) cycleInto(coords Coords, newGrid Grid) {
	// - If a cube is active and exactly 2 or 3 of its neighbors are also active,
	//     the cube remains active. Otherwise, the cube becomes inactive.
	// - If a cube is inactive but exactly 3 of its neighbors are active, the cube becomes
	//     active. Otherwise, the cube remains inactive.

	_, active := g[coords]
	var newActive bool
	activeNeighbors := g.activeNeighborCount(coords)
	if active {
		if activeNeighbors == 2 || activeNeighbors == 3 {
			newActive = true
		} else {
			newActive = false
		}
	} else {
		if activeNeighbors == 3 {
			newActive = true
		} else {
			newActive = false
		}
	}
	if newActive {
		newGrid.setActive(coords)
	}
}

func (g Grid) cycle() Grid {
	newGrid := make(Grid)

	for coords := range g {
		g.cycleInto(coords, newGrid)
		for _, neighbor := range g.getNeighbors(coords) {
			g.cycleInto(neighbor, newGrid)
		}
	}

	return newGrid
}

func day17part1(input string) any {
	deltas = make([]Coords, 0)
	for z := -1; z <= 1; z += 1 {
		for y := -1; y <= 1; y += 1 {
			for x := -1; x <= 1; x += 1 {
				if z != 0 || y != 0 || x != 0 {
					deltas = append(deltas, Coords{x, y, z, 0})
				}
			}
		}
	}

	if len(deltas) != 26 {
		panic(fmt.Sprintf("%d", len(deltas)))
	}

	grid := make(Grid)
	for y, line := range strings.Split(strings.TrimSpace(input), "\n") {
		for x, c := range []byte(line) {
			z, w := 0, 0
			coords := Coords{x, y, z, w}
			if c == '#' {
				grid.setActive(coords)
			}
		}
	}

	//grid.draw("Before any cycles:")
	grid = grid.cycle()
	//grid.draw("After 1 cycle:")
	grid = grid.cycle()
	//grid.draw("After 2 cycles:")
	grid = grid.cycle()
	//grid.draw("After 3 cycles:")
	grid = grid.cycle()
	//grid.draw("After 4 cycles:")
	grid = grid.cycle()
	//grid.draw("After 5 cycles:")
	grid = grid.cycle()
	//grid.draw("After 6 cycles:")

	return len(grid)
}

func day17part2(input string) any {
	deltas = make([]Coords, 0)
	for w := -1; w <= 1; w += 1 {
		for z := -1; z <= 1; z += 1 {
			for y := -1; y <= 1; y += 1 {
				for x := -1; x <= 1; x += 1 {
					if w != 0 || z != 0 || y != 0 || x != 0 {
						deltas = append(deltas, Coords{x, y, z, w})
					}
				}
			}
		}
	}
	if len(deltas) != 80 {
		panic(fmt.Sprintf("%d", len(deltas)))
	}

	grid := make(Grid)
	for y, line := range strings.Split(strings.TrimSpace(input), "\n") {
		for x, c := range []byte(line) {
			z, w := 0, 0
			coords := Coords{x, y, z, w}
			if c == '#' {
				grid.setActive(coords)
			}
		}
	}

	//grid.draw("Before any cycles:")
	grid = grid.cycle()
	//grid.draw("After 1 cycle:")
	grid = grid.cycle()
	//grid.draw("After 2 cycles:")
	grid = grid.cycle()
	//grid.draw("After 3 cycles:")
	grid = grid.cycle()
	//grid.draw("After 4 cycles:")
	grid = grid.cycle()
	//grid.draw("After 5 cycles:")
	grid = grid.cycle()
	//grid.draw("After 6 cycles:")

	return len(grid)
}
