package gmachine_test

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"testing/iotest"

	gmachine "github.com/bit-gophers/merit-gmachine"

	"github.com/google/go-cmp/cmp"
	"github.com/rogpeppe/go-internal/testscript"
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
	g := AssembleAndRunFromString(t, "halt")
	var want gmachine.Word = 1
	got := g.P
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestNOOP(t *testing.T) {
	t.Parallel()
	g := AssembleAndRunFromString(t, "noop halt")
	var want gmachine.Word = 2
	got := g.P
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestINCA(t *testing.T) {
	t.Parallel()
	g := AssembleAndRunFromString(t, "inca halt")
	var want gmachine.Word = 1
	got := g.A
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestDECA(t *testing.T) {
	t.Parallel()
	g := AssembleAndRunFromString(t, "SETA 2;DECA;halt")
	var want gmachine.Word = 1
	got := g.A
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestSubtract2From3Gives1(t *testing.T) {
	t.Parallel()
	g := AssembleAndRunFromFile(t, "testdata/subtract2from3.g")
	var want gmachine.Word = 1
	got := g.A
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestSETA(t *testing.T) {
	t.Parallel()
	g := AssembleAndRunFromFile(t, "testdata/setaTo5.g")
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
	g := AssembleAndRunFromFile(t, "testdata/fib.g")
	var want gmachine.Word = 89
	got := g.A
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestAssembly(t *testing.T) {
	t.Parallel()
	want := []gmachine.Word{
		gmachine.Word(gmachine.OpNOOP),
		gmachine.Word(gmachine.OpHALT),
	}
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
	AssembleAndRunFromString(t, "NOOP; halt")
}

func TestAssemblingAndRunFromFile(t *testing.T) {
	t.Parallel()
	want := []gmachine.Word{
		gmachine.Word(gmachine.OpNOOP),
		gmachine.Word(gmachine.OpHALT),
	}
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
	if err := machine.AssembleAndRunFromFile("testdata/non-existent-program.g", false); err == nil {
		t.Fatal("expected an error for invalid file")
	}
}

func TestAssemblingAndRunWithBadFile(t *testing.T) {
	t.Parallel()
	machine := gmachine.New()
	if err := machine.AssembleAndRunFromFile("testdata/invalid_program.g", false); err == nil {
		t.Fatal("expected an error for bad program file")
	}
}

func TestTokenize(t *testing.T) {
	t.Parallel()
	want := []gmachine.Token{
		{
			Kind:     gmachine.TokenInstruction,
			Value:    gmachine.Word(gmachine.OpNOOP),
			RawToken: "NOOP",
			Line:     1,
		},
		{
			Kind:     gmachine.TokenInstruction,
			Value:    gmachine.Word(gmachine.OpSETA),
			RawToken: "SETA",
			Line:     2,
		},
		{
			Kind:     gmachine.TokenNumberLiteral,
			Value:    5,
			RawToken: "5",
			Line:     2,
		},
		{
			Kind:     gmachine.TokenInstruction,
			Value:    gmachine.Word(gmachine.OpHALT),
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
	err := m.Load([]gmachine.Word{math.MaxUint64})
	if err != nil {
		t.Error(err)
	}
	err = m.Run(false)
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
	g := AssembleAndRunFromFile(t, "testdata/print_char.g")
	g.Out = buf
	want := "A"
	got := buf.String()
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestPrintHelloWorld(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	g := AssembleAndRunFromFile(t, "testdata/hello_world.g")
	g.Out = buf
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
			Col:      0,
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

func TestTokenize_RecognizeRuneLiterals(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name    string
		program string
		want    []gmachine.Token
	}
	for _, c := range []testCase{
		{
			name:    "test ascii",
			program: "'H'",
			want: []gmachine.Token{
				{
					Kind:     gmachine.TokenRuneLiteral,
					Value:    72,
					RawToken: "'H'",
					Line:     1,
					Col:      0,
				},
			},
		},
		{
			name:    "test unicode",
			program: "'л'",
			want: []gmachine.Token{
				{
					Kind:     gmachine.TokenRuneLiteral,
					Value:    gmachine.Word('л'),
					RawToken: "'л'",
					Line:     1,
					Col:      0,
				},
			},
		},
	} {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := gmachine.Tokenize(c.program)
			if err != nil {
				t.Errorf("want no error: got %v", err)
			}
			if !cmp.Equal(c.want, got) {
				t.Error(cmp.Diff(c.want, got))
			}
		})
	}
}

func TestOpCode_RequiresArgument(t *testing.T) {
	t.Parallel()
	for _, c := range []gmachine.OpCode{gmachine.OpSETA, gmachine.OpSETI} {
		if !c.RequiresArgument() {
			t.Errorf("Op code %s should require argument", c.String())
		}
	}
	for _, c := range []gmachine.OpCode{gmachine.OpNOOP, gmachine.OpHALT} {
		if c.RequiresArgument() {
			t.Errorf("Op code %s should not require argument", c.String())
		}
	}
}

func TestStateStringOutput(t *testing.T) {
	t.Parallel()
	g := AssembleAndRunFromString(t, "inca halt inca")
	want := "P: 000002 A: 000001 I: 000000 X: 000000 Y: 000000 Z: false NEXT: INCA"
	got := g.String()
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestDebugFlag(t *testing.T) {
	t.Parallel()
	output := new(bytes.Buffer)
	g := AssembleAndRunFromString(t, "inca halt")
	g.In = bytes.NewReader([]byte(""))
	g.Out = output
	got := output.String()
	if !strings.HasPrefix(got, "P:") {
		t.Errorf("Debug should start with %q", "P:")
	}
}

func TestDebugger_DecodeInstruction(t *testing.T) {
	t.Parallel()
	want := "JUMP 5"
	g := gmachine.New()
	copy(g.Memory, []gmachine.Word{gmachine.Word(gmachine.OpJUMP), 5})
	got := g.DecodeNextInstruction()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestInvertMap(t *testing.T) {
	t.Parallel()
	testMap := map[string]int{"A": 1, "B": 2, "C": 3}
	want := map[int]string{1: "A", 2: "B", 3: "C"}
	got := gmachine.InvertMap(testMap)

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestScript(t *testing.T) {
	t.Parallel()
	testscript.Run(t, testscript.Params{
		Dir: "testdata/scripts",
	})
}

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"run": gmachine.MainRun,
	}))
}

func FuzzTokenize(f *testing.F) {
	f.Add("NOOP HALT SETA 5")
	f.Fuzz(func(t *testing.T, data string) {
		got, err := gmachine.Tokenize(data)
		if len(got) == 0 && err == nil && data != " " && data != "" {
			t.Error("expected at least one token if no error is produced")
		}
	})
}

func AssembleAndRunFromFile(t *testing.T, filename string) *gmachine.Machine {
	t.Helper()

	g := gmachine.New()
	g.Out = new(bytes.Buffer)
	words, err := gmachine.AssembleFromFile(filename)

	if err != nil {
		t.Fatal(err)
	}
	err = g.Load(words)
	if err != nil {
		t.Fatal(err)
	}
	err = g.Run(false)
	if err != nil {
		t.Fatal(err)
	}

	return g
}

func AssembleAndRunFromString(t *testing.T, program string) *gmachine.Machine {
	t.Helper()

	g := gmachine.New()
	g.Out = new(bytes.Buffer)
	words, err := gmachine.Assemble(strings.NewReader(program))

	if err != nil {
		t.Fatal(err)
	}
	err = g.Load(words)
	if err != nil {
		t.Fatal(err)
	}
	err = g.Run(false)
	if err != nil {
		t.Fatal(err)
	}

	return g
}
