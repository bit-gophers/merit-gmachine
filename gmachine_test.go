package gmachine_test

import (
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
	err := g.RunProgram([]gmachine.Word{
		gmachine.OpHALT,
	})
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
	err := g.RunProgram([]gmachine.Word{
		gmachine.OpNOOP,
	})
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
	err := g.RunProgram([]gmachine.Word{
		gmachine.OpINCA,
	})
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
	err := g.RunProgram([]gmachine.Word{
		gmachine.OpSETA,
		2,
		gmachine.OpDECA,
	})
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
	err := g.RunProgram([]gmachine.Word{
		gmachine.OpSETA,
		3,
		gmachine.OpDECA,
		gmachine.OpDECA,
	})
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
	err := g.RunProgram([]gmachine.Word{
		gmachine.OpSETA,
		5,
	})
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

// 1, 2, 3, 5, 8...
func fib(n int) int {
	result := 1
	prev := 0
	for i := 1; i <= n; i++ {
		tmp := result + prev
		prev = result
		result = tmp
	}
	_ = prev
	return result
}

func TestFib2(t *testing.T) {
	t.Parallel()
	want := 8
	got := fib(5)
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestFib(t *testing.T) {
	t.Parallel()
	g := gmachine.New()
	err := g.RunProgram([]gmachine.Word{
		gmachine.OpINCA,
		gmachine.OpSETI,
		10,
		gmachine.OpMVAY,
		gmachine.OpADXY,
		gmachine.OpMVAX,
		gmachine.OpMVYA,
		gmachine.OpDECI,
		gmachine.OpJINZ,
		3,
	})
	if err != nil {
		t.Fatal(err)
	}
	var want gmachine.Word = 89
	got := g.A
	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}
