// Package gmachine implements a simple virtual CPU, known as the G-machine.
package gmachine

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// DefaultMemSize is the number of 64-bit words of memory which will be
// allocated to a new G-machine by default.
const DefaultMemSize = 1024

type OpCode Word

const (
	OpHALT OpCode = iota + 1
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
)

const (
	TokenInstruction = iota + 1
	TokenComment
	TokenNumberLiteral
	TokenRuneLiteral

	eof rune = 0
)

var kind = map[int]string{
	TokenInstruction:   "instruction",
	TokenNumberLiteral: "number literal",
}

type Word uint64

type Machine struct {
	Memory        []Word
	A, I, P, X, Y Word
	Z             bool
	Out           io.Writer
	In            io.Reader
	Debug         bool
}

func New() *Machine {
	return &Machine{
		Memory: make([]Word, DefaultMemSize),
		In:     os.Stdin,
		Out:    os.Stdout,
	}
}

func (g *Machine) Run() error {
	var inReader *bufio.Reader
	if g.Debug {
		inReader = bufio.NewReader(g.In)
	}

	for {
		if g.Debug {
			fmt.Fprint(g.Out, g.String())
			inReader.ReadLine()
		}

		op := g.Fetch()
		switch OpCode(op) {
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
			fmt.Fprintf(g.Out, "%c", g.A)
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

func (g *Machine) Peek() Word {
	op := g.Memory[g.P+1]
	return op
}

func (g *Machine) DecodeNextInstruction() string {
	opCode := OpCode(g.Memory[g.P])

	result := opCode.String()

	if opCode.RequiresArgument() {
		result += fmt.Sprintf(" %v", g.Peek())
	}

	return result
}

func (g *Machine) Load(data []Word) error {
	if len(data) > len(g.Memory) {
		return errors.New("program size exceeds memory size")
	}

	copy(g.Memory, data)
	g.P = 0
	return nil
}

func MainRun() int {
	debug := flag.Bool("debug", false, "If true print debug output")
	flag.Parse()
	g := New()
	g.Debug = *debug
	program, err := AssembleFromFile(flag.Arg(0))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}
	err = g.Load(program)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}
	err = g.Run()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}
	return 0
}

// Map of assembly instructions to OP codes
var instructions = map[string]OpCode{
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

var opCodes = InvertMap(instructions)

type Instruction struct {
	OpCode           Word
	RequiresArgument bool
}

type Token struct {
	Kind     int
	Value    Word
	RawToken string
	Line     int
	Col      int
}

func (o OpCode) RequiresArgument() bool {
	switch o {
	case OpSETA, OpSETI, OpJUMP, OpJNEQ:
		return true
	}

	return false
}

func (o OpCode) String() string {
	return opCodes[o]
}

func (t Token) String() string {
	return fmt.Sprintf("%q (%d) %s", t.RawToken, t.Value, kind[t.Kind])
}

func newToken(rawToken []rune) (Token, error) {
	stringToken := string(rawToken)
	if strings.HasPrefix(stringToken, "//") {
		return Token{
			Kind:     TokenComment,
			RawToken: stringToken,
		}, nil
	}

	tokenKind := TokenInstruction
	value, ok := instructions[strings.ToUpper(stringToken)]
	if !ok {
		if utf8.RuneCountInString(stringToken) == 3 && strings.HasPrefix(stringToken, "'") && strings.HasSuffix(stringToken, "'") {
			tokenKind = TokenRuneLiteral
			value = OpCode([]rune(stringToken)[1])
		} else {
			tokenKind = TokenNumberLiteral
			converted, err := strconv.Atoi(stringToken)
			if err != nil {
				return Token{}, fmt.Errorf("unknown instruction %q", string(rawToken))
			}
			value = OpCode(converted)
		}
	}
	return Token{
		Kind:     tokenKind,
		Value:    Word(value),
		RawToken: stringToken,
	}, nil
}

func (g *Machine) String() string {
	return fmt.Sprintf(`P: %06v A: %06v I: %06v X: %06v Y: %06v Z: %v NEXT: %v`, g.P, g.A, g.I, g.X, g.Y, g.Z, g.DecodeNextInstruction())
}

func InvertMap[K, V comparable](m map[K]V) map[V]K {
	result := map[V]K{}

	for k, v := range m {
		result[v] = k
	}

	return result
}
