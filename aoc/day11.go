package main

import (
	"strings"
)

func day11part1(input string) any {
	rows := make([][]byte, 0)
	for _, line := range strings.Split(strings.TrimSpace(input), "\n") {
		row := make([]byte, 0)
		for i := 0; i < len(line); i += 1 {
			row = append(row, line[i])
		}
		rows = append(rows, row)
	}

	rowsLen, rowLen := len(rows), len(rows[0])

	iter := 0
	for {
		changed := false
		newRows := make([][]byte, rowsLen)
		for r, row := range rows {
			newRow := make([]byte, rowLen)
			newRows[r] = newRow
			for c, char := range row {
				occupied := 0
				for dr := r - 1; dr <= r+1; dr += 1 {
					if dr == -1 || dr == rowsLen {
						continue
					}
					for dc := c - 1; dc <= c+1; dc += 1 {
						if dc == -1 || dc == rowLen || (dr == r && dc == c) {
							continue
						}
						if rows[dr][dc] == '#' {
							occupied += 1
						}
					}
				}

				newChar := char
				if char == 'L' && occupied == 0 {
					newChar = '#'
					changed = true
				}
				if char == '#' && occupied >= 4 {
					newChar = 'L'
					changed = true
				}
				newRow[c] = newChar
			}
		}

		rows = newRows
		iter = iter + 1
		if !changed {
			break
		}
	}

	count := 0
	for _, row := range rows {
		for _, char := range row {
			if char == '#' {
				count += 1
			}
		}
	}
	return count
}
