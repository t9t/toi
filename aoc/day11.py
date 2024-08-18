import fileinput
import sys

rows = sys.stdin.read().rstrip().splitlines()

while True:
    newRows = []
    changed = False
    for r, row in enumerate(rows):
        newRow = ""
        for c, char in enumerate(row):
            occupieds = 0
            for dr in range(r-1, r+2):
                if dr == -1 or dr == len(rows):
                    continue
                for dc in range(c-1, c+2):
                    if dc == -1 or dc == len(row) or (dr == r and dc == c):
                        continue
                    if rows[dr][dc] == "#":
                        occupieds += 1
            newChar = char
            if char == "L" and occupieds == 0:
                newChar = "#"
                changed = True
            if char == "#" and occupieds >= 4:
                newChar = "L"
                changed = True
            newRow += newChar
        newRows.append(newRow)

    rows = newRows
    if not changed:
        break

occupieds = 0
for row in rows:
    for char in row:
        if char == '#':
            occupieds += 1

print(occupieds)
