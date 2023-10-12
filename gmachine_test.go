package gmachine_test

import (
	"bytes"
	"math"
	"os"
	"strings"
	"testing"

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

func TestAssembleAndRunFromReader(t *testing.T) {
	t.Parallel()
	AssembleAndRunFromString(t, "NOOP; halt")
}

func TestUnknownOpCodeReturnsError(t *testing.T) {
	t.Parallel()
	m := gmachine.New()
	m.P = math.MaxUint64
	err := m.Load([]gmachine.Word{math.MaxUint64})
	if err != nil {
		t.Error(err)
	}
	err = m.Run()
	if err == nil {
		t.Error("no error")
	}
}

func TestPrintA(t *testing.T) {
	t.Parallel()
	g := AssembleAndRunFromFile(t, "testdata/print_char.g")
	want := "A"
	got := g.Out.(*bytes.Buffer).String()
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestPrintHelloWorld(t *testing.T) {
	t.Parallel()
	g := AssembleAndRunFromFile(t, "testdata/hello_world.g")
	want := "Hello World"
	got := g.Out.(*bytes.Buffer).String()
	if want != got {
		t.Errorf("want %q, got %q", want, got)
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
	g := newGMachineFromProgram(t, "inca halt")
	g.In = strings.NewReader("")
	g.Debug = true
	g.Run()
	got := g.Out.(*bytes.Buffer).String()
	if !strings.HasPrefix(got, "P:") {
		t.Errorf("Debug should start with %q got %q", "P:", got)
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

func newGMachineFromProgram(t *testing.T, program string) *gmachine.Machine {
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
	return g
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
	err = g.Run()
	if err != nil {
		t.Fatal(err)
	}

	return g
}

func AssembleAndRunFromString(t *testing.T, program string) *gmachine.Machine {
	t.Helper()
	g := newGMachineFromProgram(t, program)
	err := g.Run()
	if err != nil {
		t.Fatal(err)
	}

	return g
}
