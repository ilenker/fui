package main

import (
	"fmt"
	"time"
)

// Tests -------------------------------------------------------------------------------------
var testIndex int

func runNextTest() {
	if testIndex >= len(tests) {
		term.Println("\nDone.")
		testIndex = 0
		return
	}
	term.Clear()
	test := tests[testIndex]
	term.Println(fmt.Sprintf("[%2d/%2d]%s", testIndex+1, len(tests), test.Name))
	term.Println(fmt.Sprintf("Inputs: %s", test.Sequence))
	term.Println(fmt.Sprintf("Expect: %s", test.Expected))

	for _, char := range test.Sequence {
		calculator(char)
		time.Sleep(time.Millisecond * 100)
	}
	testIndex++
}

func runAllTests() {
	term.Println("Running tests...")
	time.Sleep(time.Second * 1)

	for _, test := range tests {
		term.Println(fmt.Sprintf("--- Test: %s [%s] ---", test.Name, test.Sequence))
		term.Println(fmt.Sprintf("--- Expected: %s ", test.Expected))

		for _, char := range test.Sequence {
			calculator(char)
			time.Sleep(time.Millisecond * 100)
		}

		// Longer pause to let you read the result
		time.Sleep(time.Second * 3)
		term.Clear()
	}
	term.Println("Done.")
}

type TestCase struct {
	Name     string
	Sequence string
	Expected string
}

var tests = []TestCase{
	{
		Name:     "Empty / Just Equals",
		Sequence: "＝",
		Expected: "blank",
	},
	{
		Name:     "Operator First",
		Sequence: "＋5＝",
		Expected: "5",
	},
	{
		Name:     "Basic Addition",
		Sequence: "100＋55＝",
		Expected: "155",
	},
	{
		Name:     "Decimal Subtraction",
		Sequence: "5.5－.5＝",
		Expected: "5",
	},
	{
		Name:     "Multiplication with Zeros",
		Sequence: "100＊0＝",
		Expected: "0",
	},
	{
		Name:     "Division (Integer Result)",
		Sequence: "12／4＝",
		Expected: "3",
	},
	{
		Name:     "Implicit Reset on Number",
		Sequence: "1＋1＝9＋9＝",
		Expected: "18",
	},
	{
		Name:     "Implicit Reset on Period",
		Sequence: "2＋2＝.5＋.5＝",
		Expected: "1",
	},
	{
		Name:     "Repeated Addition",
		Sequence: "1＋1＝＝＝",
		Expected: "4",
	},
	{
		Name:     "Repeated Division",
		Sequence: "100／2＝＝＝",
		Expected: "12.5",
	},
	{
		Name:     "Repeated Multiplication",
		Sequence: "2＊2＝＝＝",
		Expected: "16",
	},
	{
		Name:     "Operation Chaining",
		Sequence: "5＋5＝＊2＝－5＝",
		Expected: "5",
	},
	{
		Name:     "Implied Leading Zeros",
		Sequence: ".1＋.1＝",
		Expected: "0.2",
	},
	{
		Name:     "Many Zeros",
		Sequence: "0000＋0001＝",
		Expected: "1",
	},
	{
		Name:     "Division by Zero",
		Sequence: "5／0＝",
		Expected: "No idea",
	},
	{
		Name:     "Overwrite Op",
		Sequence: "1＋－＊5＝",
		Expected: "5",
	},
}
