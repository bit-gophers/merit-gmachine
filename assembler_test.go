package gmachine_test

import (
	"fmt"
	"strings"
	"testing"
	"testing/iotest"

	gmachine "github.com/bit-gophers/merit-gmachine"
	"github.com/google/go-cmp/cmp"
)

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

func TestLabels(t *testing.T) {
	t.Parallel()
	want := []gmachine.Word{gmachine.Word(gmachine.OpJUMP), 2, 0}
	got, err := gmachine.Assemble(strings.NewReader("JUMP main;main:"))
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestAssembleFromFile(t *testing.T) {
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

func TestAssembleErrorReaderReturnsError(t *testing.T) {
	t.Parallel()
	reader := iotest.ErrReader(fmt.Errorf("some error"))
	_, err := gmachine.Assemble(reader)
	if err == nil {
		t.Error("no error")
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

func TestTokenizeLog(t *testing.T) {
	t.Parallel()

	tokenizer := gmachine.NewTokenizer()
	_, err := tokenizer.Run("//foo")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(tokenizer.Log.String(), "emit") {
		t.Fatal("expected tokenizer log to contain \"emit\"")
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

func FuzzTokenize(f *testing.F) {
	f.Add("NOOP HALT SETA 5")
	f.Fuzz(func(t *testing.T, data string) {
		got, err := gmachine.Tokenize(data)
		if len(got) == 0 && err == nil && data != " " && data != "" {
			t.Error("expected at least one token if no error is produced")
		}
	})
}
