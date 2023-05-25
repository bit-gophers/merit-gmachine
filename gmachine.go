// Package gmachine implements a simple virtual CPU, known as the G-machine.
package gmachine

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// DefaultMemSize is the number of 64-bit words of memory which will be
// allocated to a new G-machine by default.
const DefaultMemSize = 1024

const (
	OpHALT = iota + 1
	OpNOOP
	OpINCA
	OpDECA
	OpSETA
	OpSETI
	OpDECI
	OpJINZ
	OpMVAY
	OpADXY
	OpMVAX
	OpMVYA
	OpOUTA
	OpJUMP
	OpINCI
	OpLDAI
	OpCMPI
	OpJNEQ

	TokenInstruction = iota
	TokenArgument
	TokenComment

	eof rune = 0
)

var kind = map[int]string{
	TokenInstruction: "instruction",
	TokenArgument:    "argument",
}

type Word uint64

type Machine struct {
	Memory        []Word
	A, I, P, X, Y Word
	Z             bool
	out           io.Writer
}

func New() *Machine {
	return &Machine{
		Memory: make([]Word, DefaultMemSize),
		out:    os.Stdout,
	}
}

func NewWithOutput(out io.Writer) *Machine {
	g := New()
	g.out = out
	return g
}

func (g *Machine) Run() error {
	for {
		// fmt.Printf("P: %d NextOp: %d A: %d I: %d X: %d Y: %d\n", g.P, g.Memory[g.P], g.A, g.I, g.X, g.Y)
		op := g.Fetch()
		switch op {
		case OpHALT:
			return nil
		case OpNOOP:
		case OpINCA:
			g.A++
		case OpDECA:
			g.A--
		case OpSETA:
			g.A = g.Fetch()
		case OpSETI:
			g.I = g.Fetch()
		case OpDECI:
			g.I--
		case OpJINZ:
			if g.I != 0 {
				g.P = g.Fetch()
			} else {
				g.P++
			}
		case OpMVAY:
			g.Y = g.A
		case OpADXY:
			g.Y += g.X
		case OpMVAX:
			g.X = g.A
		case OpMVYA:
			g.A = g.Y
		case OpOUTA:
			fmt.Fprintf(g.out, "%c", g.A)
		case OpJUMP:
			g.P = g.Fetch()
		case OpINCI:
			g.I++
		case OpLDAI:
			g.A = g.Memory[g.I+g.Fetch()]
		case OpCMPI:
			g.Z = g.I == g.Fetch()
		case OpJNEQ:
			if !g.Z {
				g.P = g.Fetch()
			}
		default:
			return fmt.Errorf("unknown opcode %d", op)
		}
	}
}

func (g *Machine) Fetch() Word {
	op := g.Memory[g.P]
	g.P++
	return op
}

func (g *Machine) RunProgram(data []Word) error {
	copy(g.Memory, data)
	g.P = 0
	return g.Run()
}

// Map of assembly instructions to OP codes
var instructions = map[string]Word{
	"ADXY": OpADXY,
	"DECA": OpDECA,
	"DECI": OpDECI,
	"HALT": OpHALT,
	"INCA": OpINCA,
	"JINZ": OpJINZ,
	"MVAX": OpMVAX,
	"MVAY": OpMVAY,
	"MVYA": OpMVYA,
	"NOOP": OpNOOP,
	"OUTA": OpOUTA,
	"SETA": OpSETA,
	"SETI": OpSETI,
	"JUMP": OpJUMP,
	"INCI": OpINCI,
	"LDAI": OpLDAI,
	"CMPI": OpCMPI,
	"JNEQ": OpJNEQ,
}

type Token struct {
	Kind     int
	Value    Word
	RawToken string
	Line     int
	Col      int
}

func (t Token) String() string {
	return fmt.Sprintf("%q (%d) %s", t.RawToken, t.Value, kind[t.Kind])
}

func Tokenize(data []rune) ([]Token, error) {
	t := new(tokenizer)
	t.line = 1

	if os.Getenv("TOKENIZE_LOGS") != "" {
		t.debug = os.Stderr
	} else {
		t.debug = io.Discard
	}

	t.input = data
	for state := wantToken; state != nil; {
		state = state(t)
		if t.err != nil {
			return nil, t.err
		}
	}
	return t.result, nil
}

type tokenizer struct {
	sourceName       string
	debug            io.Writer
	input            []rune
	start, pos, line int
	result           []Token
	err              error
}

func (t *tokenizer) next() rune {
	next := t.peek()
	if next != eof {
		t.pos++
	}
	return next
}

func (t *tokenizer) peek() rune {
	if t.pos >= len(t.input) {
		return eof
	}
	next := t.input[t.pos]
	return next
}

func (t *tokenizer) skip() {
	t.start = t.pos
}

func (t *tokenizer) backup() {
	t.pos--
}

func newToken(rawToken []rune) (Token, error) {
	stringToken := string(rawToken)
	if strings.HasPrefix(stringToken, "//") {
		return Token{
			Kind:     TokenComment,
			RawToken: stringToken,
		}, nil
	}

	caseInsensitiveToken := strings.ToUpper(stringToken)
	tokenKind := TokenInstruction
	value, ok := instructions[caseInsensitiveToken]
	if !ok {
		tokenKind = TokenArgument
		converted, err := strconv.Atoi(caseInsensitiveToken)
		if err != nil {
			return Token{}, fmt.Errorf("unknown instruction %q", string(rawToken))
		}
		value = Word(converted)
	}
	return Token{
		Kind:     tokenKind,
		Value:    value,
		RawToken: caseInsensitiveToken,
	}, nil
}

func (t *tokenizer) emit() {
	token, err := newToken(t.input[t.start:t.pos])
	if err != nil {
		t.err = fmt.Errorf("%d: syntax error: %w", t.line, err)
	}
	token.Line = t.line
	t.log("emit", token)
	t.result = append(t.result, token)
}

func (t *tokenizer) log(args ...interface{}) {
	fmt.Fprintln(t.debug, args...)
}

func (t *tokenizer) logState(stateName string) {
	next := "EOF"
	if t.pos < len(t.input) {
		next = string(t.input[t.pos])
	}
	t.log(fmt.Sprintf("%s: [%s] -> %s",
		stateName,
		string(t.input[t.start:t.pos]),
		next,
	))
}

type stateFunc func(*tokenizer) stateFunc

func wantToken(t *tokenizer) stateFunc {
	for {
		t.logState("wantToken")
		switch t.next() {
		case '\n':
			t.line++
			t.skip()
		case ' ', ';':
			t.skip()
		case eof:
			return nil
		default:
			t.backup()
			return inToken
		}
	}
}

func inToken(t *tokenizer) stateFunc {
	for {
		t.logState("inToken")
		switch t.next() {
		case '/':
			if t.peek() == '/' {
				return inComment
			}
			t.err = fmt.Errorf("%d: syntax error: expected '/' got '%c'", t.line, t.peek())
			return nil
		case '\n', ' ', ';':
			t.backup()
			t.emit()
			return wantToken
		case eof:
			t.emit()
			return nil
		}
	}
}

func inComment(t *tokenizer) stateFunc {
	for {
		t.logState("inComment")
		switch t.next() {
		case '\n':
			t.backup()
			t.emit()
			return wantToken
		case eof:
			t.emit()
			return nil
		}
	}
}

func Assemble(input io.Reader) ([]Word, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	var program []Word
	tokens, err := Tokenize([]rune(string(data)))
	if err != nil {
		return nil, err
	}
	argRequired := false
	for _, t := range tokens {
		if t.Kind == TokenComment {
			continue
		}

		if t.Kind == TokenInstruction && argRequired {
			return nil, fmt.Errorf("line %d: unexpected instruction %q", t.Line, t.RawToken)
		}
		switch t.RawToken {
		case "SETA", "SETI", "JUMP":
			argRequired = true
		default:
			argRequired = false
		}
		program = append(program, t.Value)
	}
	program = append(program, OpHALT)
	return program, nil
}

func AssembleFromFile(filename string) ([]Word, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	program, err := Assemble(file)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", filename, err)
	}
	return program, nil
}

func (g *Machine) AssembleAndRunFromString(program string) error {
	words, err := Assemble(strings.NewReader(program))
	if err != nil {
		return err
	}
	return g.RunProgram(words)
}

func (g *Machine) AssembleAndRunFromFile(filename string) error {
	program, err := AssembleFromFile(filename)
	if err != nil {
		return err
	}
	return g.RunProgram(program)
}
