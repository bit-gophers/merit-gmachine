package gmachine_test

import (
	"strings"
	"testing"

	"gmachine"

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
	err := g.AssembleAndRunFromString("noop")
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
	err := g.AssembleAndRunFromString("inca")
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
	err := g.AssembleAndRunFromString("SETA 2;DECA")
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
	want := []gmachine.Word{gmachine.OpNOOP, gmachine.OpHALT, gmachine.OpHALT}
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
	want := []gmachine.Word{gmachine.OpNOOP, gmachine.OpHALT, gmachine.OpHALT}
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
	got, err := gmachine.Tokenize([]rune("NOOP\nSETA 5 \n HALT"))
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestTokenizeError(t *testing.T) {
	t.Parallel()
	program := []rune("[")
	_, err := gmachine.Tokenize(program)
	if err == nil {
		t.Error("want error: got nil")
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
