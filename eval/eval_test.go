// eval_test.go - Simple test-cases for our evaluator.

package eval

import (
	"strings"
	"testing"

	"github.com/skx/gobasic/object"
	"github.com/skx/gobasic/tokenizer"
)

// TestTrace tests that getting/setting the tracing-flag works as expected.
func TestTrace(t *testing.T) {
	input := `10 PRINT "OK\n"`
	tokener := tokenizer.New(input)
	e, _ := New(tokener)

	// Tracing is off by default
	if e.GetTrace() != false {
		t.Errorf("tracing should not be enabled by default")
	}

	// Enable tracing
	e.SetTrace(true)

	// Tracing is now on
	if e.GetTrace() != true {
		t.Errorf("tracing should have been enabled, but was not")
	}
}

// TestVariables gets/sets some variables and ensures they work
func TestVariables(t *testing.T) {

	type Test struct {
		Name   string
		Object object.Object
	}

	var vars []Test

	//
	// Setup some test variables.
	//
	vars = append(vars, Test{Name: "number", Object: object.Number(33)})
	vars = append(vars, Test{Name: "string", Object: object.String("Steve")})
	vars = append(vars, Test{Name: "error", Object: object.Error("Blah")})

	//
	// Test getting/setting each variable.
	//
	for _, v := range vars {

		input := `10 PRINT "OK\n"`
		tokener := tokenizer.New(input)
		e, _ := New(tokener)

		//
		// By default the variable won't exist.
		//
		cur := e.GetVariable(v.Name)
		if cur.Type() != object.ERROR {
			t.Errorf("Unexpectedly managed to retrieve a missing variable")
		}

		//
		// Set it
		//
		e.SetVariable(v.Name, v.Object)

		//
		// Ensure it was set
		//
		cur = e.GetVariable(v.Name)
		if cur.Type() != v.Object.Type() {
			t.Errorf("Retrieved variable '%s' had the wrong type %s != %s", v.Name, cur.Type(), v.Object.Type())
		}
	}
}

// TestData tests that invalid data items cause the program to fail.
func TestData(t *testing.T) {
	type Test struct {
		Input string
		Valid bool
	}

	vars := []Test{{Input: `10 DATA 2,1,2`, Valid: true},
		{Input: `10 DATA "2","1","2"`, Valid: true},
		{Input: `10 DATA "2","steve",2
`, Valid: true},
		{Input: `10 DATA LET, b, c`, Valid: false},
	}

	//
	// Test reading each set of data.
	//
	for _, v := range vars {

		tokener := tokenizer.New(v.Input)
		_, err := New(tokener)

		if v.Valid {
			if err != nil {
				t.Errorf("Expected error, received one: %s!", err.Error())
			}
		} else {
			if err == nil {
				t.Errorf("Expected error, received none for input %s", v.Input)
			}
		}
	}
}

// TestCompare tests our comparison operation, via IF
func TestCompare(t *testing.T) {
	type Test struct {
		Input string
		Var   string
		Val   float64
	}

	tests := []Test{
		{Input: `10 IF 1 < 10 THEN LET a=1 ELSE LET a=0`, Var: "a", Val: 1},
		{Input: `20 IF 1 <= 10 THEN LET b=1 ELSE LET b=2`, Var: "b", Val: 1},
		{Input: `10 IF 11 > 7 THEN let c=1 ELSE LET c=0`, Var: "c", Val: 1},
		{Input: `40 IF 11 >= 7 THEN let d=1 ELSE LET d=3`, Var: "d", Val: 1},
		{Input: `50 IF 1 = 1 THEN let e=1 ELSE LET e=3`, Var: "e", Val: 1},
		{Input: `60 IF 1 <> 3 THEN let f=13 ELSE LET f=3`, Var: "f", Val: 13},
		{Input: `70 IF 1 <> 1 THEN let g=3 ELSE LET g=33`, Var: "g", Val: 33},
		{Input: `80 IF "a" < "b" THEN LET A=1 ELSE LET A=0`, Var: "A", Val: 1},
		{Input: `90 IF "a" <= "a" THEN LET B=1 ELSE LET B=2`, Var: "B", Val: 1},
		{Input: `100 IF "b" > "a" THEN let C=1 ELSE LET C=0`, Var: "C", Val: 1},
		{Input: `110 IF "c" >= "a" THEN let D=1 ELSE LET D=3`, Var: "D", Val: 1},
		{Input: `120 IF "moi" = "moi" THEN let E=1 ELSE LET E=3`, Var: "E", Val: 1},
		{Input: `130 IF "steve" <> "kemp" THEN let F=13 ELSE LET F=3`, Var: "F", Val: 13},
		{Input: `140 IF "a" <> "a" THEN let G=3 ELSE LET G=33`, Var: "G", Val: 33},
		{Input: `10 LET a=1
20 IF a THEN LET t=11 ELSE let t=10
`, Var: "t", Val: 11},
		{Input: `10 LET a=0
20 IF a THEN LET t=11 ELSE let t=10
`, Var: "t", Val: 10},
		{Input: `10 LET a="steve"
20 IF a THEN LET tt=11 ELSE let tt=10
`, Var: "tt", Val: 11},
		{Input: `10 LET a=""
20 IF a THEN LET tt=11 ELSE let tt=10
`, Var: "tt", Val: 10},
	}

	//
	// Test each comparison
	//
	for _, v := range tests {

		tokener := tokenizer.New(v.Input + "\n")
		e, err := New(tokener)
		if err != nil {
			t.Errorf("Error parsing %s - %s", v.Input, err.Error())
		}

		e.Run()

		//
		// By default the variable won't exist.
		//
		cur := e.GetVariable(v.Var)
		if cur.Type() == object.ERROR {
			t.Errorf("Variable %s does not exist", v.Var)
		}
		if cur.Type() != object.NUMBER {
			t.Errorf("Variable %s had wrong type: %s", v.Var, cur.String())
		}
		out := cur.(*object.NumberObject).Value
		if out != v.Val {
			t.Errorf("Expected %s to be %f, got %f", v.Var, v.Val, out)
		}
	}
}

// TestMismatchedTypes tests that expr() errors on mismatched types.
func TestMismatchedTypes(t *testing.T) {
	input := `10 LET a=3
20 LET b="steve"
30 LET c = a + b
`
	tokener := tokenizer.New(input)
	e, err := New(tokener)
	if err != nil {
		t.Errorf("Error parsing %s - %s", input, err.Error())
	}

	err = e.Run()

	if err == nil {
		t.Errorf("Expected to see an error, but didn't.")
	}
	if !strings.Contains(err.Error(), "type mismatch") {
		t.Errorf("Our error-message wasn't what we expected")
	}
}

// TestMismatchedTypesTerm tests that term() errors on mismatched types.
func TestMismatchedTypesTerm(t *testing.T) {
	input := `10 LET a="steve"
20 LET b = ( a * 2 ) + ( a * 33 )
`
	tokener := tokenizer.New(input)
	e, err := New(tokener)
	if err != nil {
		t.Errorf("Error parsing %s - %s", input, err.Error())
	}

	err = e.Run()

	if err == nil {
		t.Errorf("Expected to see an error, but didn't.")
	}
	if !strings.Contains(err.Error(), "handles integers") {
		t.Errorf("Our error-message wasn't what we expected")
	}
}

// TestStringFail tests that expr() errors on bogus string operations.
func TestStringFail(t *testing.T) {
	input := `10 LET a="steve"
20 LET b="steve"
30 LET c = a - b
`
	tokener := tokenizer.New(input)
	e, err := New(tokener)
	if err != nil {
		t.Errorf("Error parsing %s - %s", input, err.Error())
	}

	err = e.Run()

	if err == nil {
		t.Errorf("Expected to see an error, but didn't.")
	}
	if !strings.Contains(err.Error(), "not supported for strings") {
		t.Errorf("Our error-message wasn't what we expected")
	}
}

// TestExprTerm tests that expr() errors on unclosed brackets.
func TestExprTerm(t *testing.T) {
	input := `10 LET a = ( 3 + 3 * 33
20 PRINT a "\n"
`
	tokener := tokenizer.New(input)
	e, err := New(tokener)
	if err != nil {
		t.Errorf("Error parsing %s - %s", input, err.Error())
	}

	err = e.Run()

	if err == nil {
		t.Errorf("Expected to see an error, but didn't.")
	}
	if !strings.Contains(err.Error(), "Unclosed bracket") {
		t.Errorf("Our error-message wasn't what we expected")
	}
}

// Test that our bounds-checking of the input-program works.
//
// The bounds-checks are here largely as a result of the fuzz-testing,
// as described in `FUZZING.md`.
func TestEOF(t *testing.T) {

	tests := []string{
		"10 LET a = 3 *",
		"20 LET a = ( 3 * 3 ) + ",
		"30 LET a = ( 3",
		"40 LET a = (",
		"50 LET a = 3 * 3 / 3 +",
		"60 GOSUB",
		"70 GOTO",
		"80 INPUT",
		"90 INPUT \"test\"",
		"100 INPUT \"test\", ",
		"100 LET",
		"110 LET x",
		"120 LET x=",
		"130 NEXT",

		"10 PRINT 3 +",
		"10 PRINT 3 /",
		"10 PRINT 3 *",
		"10 IF 3 ",
		"10 IF \"steve\" ",
		"10 IF  ",
		"10 FOR I = 1 TO 3 STEP",
		"10 FOR I = 1 TO ",
		"10 FOR I = 1 ",

		// multi-line tests:
		`140 DATA 3,4,5
150 READ`,
	}
	for _, test := range tests {

		tokener := tokenizer.New(test)
		e, err := New(tokener)
		if err != nil {
			t.Errorf("Error parsing %s - %s", test, err.Error())
		}

		err = e.Run()
		if err == nil {
			t.Errorf("Expected error running '%s', got none", test)
		}
		if !strings.Contains(err.Error(), "end of program") {
			t.Errorf("Error '%s' wasn't an end-of-program error!", err.Error())
		}
	}
}

// TestMaths tests addition, subtraction, multiplication, division, etc.
func TestMaths(t *testing.T) {
	type Test struct {
		Input  string
		Result float64
	}

	tests := []Test{
		{Input: "3 + 3", Result: 6},
		{Input: "3 - 1", Result: 2},
		{Input: "6 / 2", Result: 3},
		{Input: "6 * 5", Result: 30},
		{Input: "2 ^ 3", Result: 8},
		{Input: "4 % 2", Result: 0},
	}

	for _, test := range tests {

		tokener := tokenizer.New("LET x =" + test.Input + "\n")
		e, err := New(tokener)
		if err != nil {
			t.Errorf("Error parsing %s - %s", test.Input, err.Error())
		}

		e.Run()

		cur := e.GetVariable("x")
		if cur.Type() == object.ERROR {
			t.Errorf("Variable x does not exist!")
		}
		if cur.Type() != object.NUMBER {
			t.Errorf("Variable x had wrong type: %s", cur.String())
		}
		out := cur.(*object.NumberObject).Value
		if out != test.Result {
			t.Errorf("Expected x to be %f, got %f", test.Result, out)
		}
	}
}

// TestRead ensures that the READ statement is sane.
func TestRead(t *testing.T) {

	//
	// This will fail because READ requires an ident.
	//
	fail1 := `
10 DATA "foo", "bar", "baz"
20 READ 3
`

	e, err := FromString(fail1)
	if err != nil {
		t.Errorf("Error parsing %s - %s", fail1, err.Error())
	}
	err = e.Run()
	if err == nil {
		t.Errorf("Expected to see an error, but didn't.")
	}
	if !strings.Contains(err.Error(), "Expected identifier") {
		t.Errorf("Our error-message wasn't what we expected")
	}

	//
	// This will fail because we READ too far.
	//
	fail2 := `
10 DATA "a", "b", "c"
20 READ a, b, c, d
`
	e, err = FromString(fail2)
	if err != nil {
		t.Errorf("Error parsing %s - %s", fail2, err.Error())
	}
	err = e.Run()
	if err == nil {
		t.Errorf("Expected to see an error, but didn't.")
	}
	if !strings.Contains(err.Error(), "Read past the end of our DATA storage") {
		t.Errorf("Our error-message wasn't what we expected")
	}

	//
	// Now a working example.
	//
	ok1 := `
10 DATA "Cat", "Kissa"
20 READ a
`
	e, err = FromString(ok1)
	if err != nil {
		t.Errorf("Error parsing %s - %s", ok1, err.Error())
	}
	err = e.Run()
	if err != nil {
		t.Errorf("Expected no error, but found one: %s", err.Error())
	}

	//
	// Now we should be able to validate our read succeeded.
	//
	out := e.GetVariable("a")
	if out.Type() != object.STRING {
		t.Errorf("Variable %s had wrong type: %s", "a", out.String())
	}
	val := out.(*object.StringObject).Value
	if val != "Cat" {
		t.Errorf("Expected %s to be %s, got %s", "a", "Cat", val)
	}

	//
	// Now a "working" example.
	//
	ok2 := `
10 DATA "Cat", "Kissa"
20 READ ,,,,,,,,,,,,,,,,,,,,,,
`
	e, err = FromString(ok2)
	if err != nil {
		t.Errorf("Error parsing %s - %s", ok2, err.Error())
	}
	err = e.Run()
	if err != nil {
		t.Errorf("Expected no error, but found one: %s", err.Error())
	}

}
