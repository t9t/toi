package main

import (
	"os"
	"testing"
)

func TestToi(t *testing.T) {
	cases := []struct {
		Filename string
		Stdin    string
	}{
		{"arrays", ""},
		{"assignment", ""},
		{"builtinFuncs", "10\n20"},
		{"comment", ""},
		{"conditionals", ""},
		{"for", ""},
		{"if", ""},
		{"inputLines", "asdf\nkek"},
		{"logicalOperators", ""},
		{"loops", ""},
		{"maps", ""},
		{"math", ""},
		{"printNumbers", ""},
		{"strings", ""},
		{"while", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.Filename, func(t *testing.T) {
			baseFilename := "toi/" + testCase.Filename
			expectedBytes, err := os.ReadFile(baseFilename + ".out")
			if err != nil {
				t.Errorf("error readint out file for '%s': %v", testCase.Filename, err)
				t.FailNow()
				return
			}
			expected := string(expectedBytes)

			stdout, err := runScriptFile(baseFilename+".toi", testCase.Stdin)
			if err != nil {
				t.Errorf("expected no error but got: %v", err)
				t.Fail()
			} else if stdout != expected {
				t.Errorf("output not as expected; expected:\n###%s###\nactual:\n###%s###", expected, stdout)
				t.Fail()
			}
		})
	}
}
