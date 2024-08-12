package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestAoc(t *testing.T) {
	outputsData, err := os.ReadFile("aoc/outputs.txt")
	if err != nil {
		t.Fatal(err)
	}
	outputs := strings.Split(strings.TrimSpace(string(outputsData)), "\n")

	day := 1
	part := 1

	for _, expected := range outputs {
		expected = expected + "\n"
		t.Run(fmt.Sprintf("aoc day %d part %d", day, part), func(t *testing.T) {
			inputData, err := os.ReadFile(fmt.Sprintf("../aoc/input/2020/%d.txt", day))
			if err != nil {
				t.Fatal(err)
			}

			stdout, err := runScriptFile(fmt.Sprintf("aoc/2020.%02d.%d.toi", day, part), string(inputData))
			if err != nil {
				t.Errorf("expected no error but got: %v", err)
				t.Fail()
			} else if stdout != expected {
				t.Errorf("output not as expected; expected:\n###%s###\nactual:\n###%s###", expected, stdout)
				t.Fail()
			}
		})

		if part == 2 {
			part = 1
			day += 1
		} else {
			part += 1
		}
	}
}
