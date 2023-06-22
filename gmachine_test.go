package gmachine_test

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"
	"testing/iotest"

	gmachine "github.com/bit-gophers/merit-gmachine"

	"github.com/google/go-cmp/cmp"
)

func TestNew(t *testing.T) {
	t.Parallel()
	g := gmachine.New()
	wantMemSize := gmachine.DefaultMemSize
	gotMemSize := len(g.Memory)
	if wantMemSize != gotMemSize {
		t.Errorf("want %d words of memory, got %d", wantMemSize, gotMemSize)
	}
	var wantP gmachine.Word = 0
	if wantP != g.P {
		t.Errorf("want initial P value %d, got %d", wantP, g.P)
	}
	var wantA gmachine.Word = 0
	if wantA != g.A {
		t.Errorf("want initial P value %d, got %d", wantA, g.A)
	}

	var wantMemValue gmachine.Word = 0
	gotMemValue := g.Memory[gmachine.DefaultMemSize-1]
	if wantMemValue != gotMemValue {
		t.Errorf("want last memory location to contain %d, got %d", wantMemValue, gotMemValue)
	}
}

func TestHALT(t *testing.T) {
	t.Parallel()
	g := gmachine.New()
	err := g.AssembleAndRunFromString("halt")
	if err != nil {
		t.Fatal(err)
	}
	var want gmachine.Word = 1
	got := g.P
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestNOOP(t *testing.T) {
	t.Parallel()
	g := gmachine.New()
	err := g.AssembleAndRunFromString("noop halt")
	if err != nil {
		t.Fatal(err)
	}
	var want gmachine.Word = 2
	got := g.P
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestINCA(t *testing.T) {
	t.Parallel()
	g := gmachine.New()
	err := g.AssembleAndRunFromString("inca halt")
	if err != nil {
		t.Fatal(err)
	}
	var want gmachine.Word = 1
	got := g.A
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestDECA(t *testing.T) {
	t.Parallel()
	g := gmachine.New()
	err := g.AssembleAndRunFromString("SETA 2;DECA;halt")
	if err != nil {
		t.Fatal(err)
	}
	var want gmachine.Word = 1
	got := g.A
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestSubtract2From3Gives1(t *testing.T) {
	t.Parallel()
	g := gmachine.New()
	err := g.AssembleAndRunFromFile("testdata/subtract2from3.g")
	if err != nil {
		t.Fatal(err)
	}
	var want gmachine.Word = 1
	got := g.A
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestSETA(t *testing.T) {
	t.Parallel()
	g := gmachine.New()
	err := g.AssembleAndRunFromFile("testdata/setaTo5.g")
	if err != nil {
		t.Fatal(err)
	}
	var want gmachine.Word = 5
	got := g.A
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
	var wantP gmachine.Word = 3
	gotP := g.P
	if wantP != gotP {
		t.Error(cmp.Diff(wantP, gotP))
	}
}

func TestFib(t *testing.T) {
	t.Parallel()
	g := gmachine.New()
	err := g.AssembleAndRunFromFile("testdata/fib.g")
	if err != nil {
		t.Fatal(err)
	}
	var want gmachine.Word = 89
	got := g.A
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestAssembly(t *testing.T) {
	t.Parallel()
	want := []gmachine.Word{gmachine.OpNOOP, gmachine.OpHALT}
	got, err := gmachine.Assemble(strings.NewReader("NOOP; halt"))
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestAssembleAndRunFromReader(t *testing.T) {
	t.Parallel()
	machine := gmachine.New()
	if err := machine.AssembleAndRunFromString("NOOP; halt"); err != nil {
		t.Fatal(err)
	}
}

func TestAssemblingAndRunFromFile(t *testing.T) {
	t.Parallel()
	want := []gmachine.Word{gmachine.OpNOOP, gmachine.OpHALT}
	got, err := gmachine.AssembleFromFile("testdata/halting_program.g")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestAssemblingAndRunWithNonExistentFile(t *testing.T) {
	t.Parallel()
	machine := gmachine.New()
	if err := machine.AssembleAndRunFromFile("testdata/non-existent-program.g"); err == nil {
		t.Fatal("expected an error for invalid file")
	}
}

func TestAssemblingAndRunWithBadFile(t *testing.T) {
	t.Parallel()
	machine := gmachine.New()
	if err := machine.AssembleAndRunFromFile("testdata/invalid_program.g"); err == nil {
		t.Fatal("expected an error for bad program file")
	}
}

func FuzzTokenize(f *testing.F) {
	f.Add("NOOP HALT SETA 5")
	f.Fuzz(func(t *testing.T, data string) {
		//machine := gmachine.New()
		//_ = machine.AssembleAndRunFromString(data)

		got, err := gmachine.Tokenize(data)
		if len(got) == 0 && err == nil && data != " " && data != "" {
			t.Error("expected at least one token if no error is produced")
		}
	})
}

func TestTokenize(t *testing.T) {
	t.Parallel()
	want := []gmachine.Token{
		{
			Kind:     gmachine.TokenInstruction,
			Value:    gmachine.OpNOOP,
			RawToken: "NOOP",
			Line:     1,
		},
		{
			Kind:     gmachine.TokenInstruction,
			Value:    gmachine.OpSETA,
			RawToken: "SETA",
			Line:     2,
		},
		{
			Kind:     gmachine.TokenArgument,
			Value:    5,
			RawToken: "5",
			Line:     2,
		},
		{
			Kind:     gmachine.TokenInstruction,
			Value:    gmachine.OpHALT,
			RawToken: "HALT",
			Line:     3,
		},
	}
	got, err := gmachine.Tokenize("NOOP\nSETA 5 \n HALT\n")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestTokenizeError(t *testing.T) {
	t.Parallel()
	program := "["
	_, err := gmachine.Tokenize(program)
	if err == nil {
		t.Error("want error: got nil")
	}
}

func TestTokenizeComment(t *testing.T) {
	t.Parallel()
	program := "//Hello"
	_, err := gmachine.Tokenize(program)
	if err != nil {
		t.Errorf("want no error: got %v", err)
	}
}

func TestErrorForBogusInstruction(t *testing.T) {
	t.Parallel()
	_, err := gmachine.AssembleFromFile("testdata/syntax_error.g")
	if err == nil {
		t.Fatal("want error for bogus instruction, got nil")
	}
}

func TestSyntaxErrorOnLine(t *testing.T) {
	t.Parallel()
	wantPrefix := `testdata/syntax_error_line_2.g:2:`
	_, err := gmachine.AssembleFromFile("testdata/syntax_error_line_2.g")
	if !strings.HasPrefix(err.Error(), wantPrefix) {
		t.Error("want prefix", wantPrefix, "got", err.Error())
	}
}

func TestAssemblyError(t *testing.T) {
	t.Parallel()
	_, err := gmachine.AssembleFromFile("testdata/assembly_error.g")
	if err == nil {
		t.Fatal("Should have got error from file")
	}
}

func TestUnknownOpCodeReturnsError(t *testing.T) {
	t.Parallel()
	m := gmachine.New()
	m.P = math.MaxUint64
	err := m.RunProgram([]gmachine.Word{math.MaxUint64})
	if err == nil {
		t.Error("no error")
	}
}

func TestAssembleErrorReaderReturnsError(t *testing.T) {
	t.Parallel()
	reader := iotest.ErrReader(fmt.Errorf("some error"))
	_, err := gmachine.Assemble(reader)
	if err == nil {
		t.Error("no error")
	}
}

func TestPrintA(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	g := gmachine.NewWithOutput(buf)
	err := g.AssembleAndRunFromFile("testdata/print_char.g")
	if err != nil {
		t.Fatal(err)
	}
	want := "A"
	got := buf.String()
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestPrintHelloWorld(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	g := gmachine.NewWithOutput(buf)
	err := g.AssembleAndRunFromFile("testdata/hello_world.g")
	if err != nil {
		t.Fatal(err)
	}
	want := "Hello World"
	got := buf.String()
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestTokenize_RecognizeNumberLiterals(t *testing.T) {
	t.Parallel()
	program := "15"
	want := []gmachine.Token{
		{
			Kind:     gmachine.TokenNumberLiteral,
			Value:    15,
			RawToken: "15",
			Line:     1,
			Col:      1,
		},
	}
	got, err := gmachine.Tokenize(program)
	if err != nil {
		t.Errorf("want no error: got %v", err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestToken_RequiresArgument(t *testing.T) {
	t.Parallel()
	token := gmachine.Token{
		Kind:  gmachine.TokenInstruction,
		Value: gmachine.OpSETA,
	}
	if !token.RequiresArgument() {
		t.Error("token should require argument")
	}
}
