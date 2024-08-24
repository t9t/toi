#!/usr/bin/env bash

set -eu

# see aoc_test.go
export TEST_AOC="${1:-}"

go vet
go test -count=1 ./...
