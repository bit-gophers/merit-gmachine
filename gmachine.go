// Package gmachine implements a simple virtual CPU, known as the G-machine.
package gmachine

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// DefaultMemSize is the number of 64-bit words of memory which will be
// allocated to a new G-machine by default.
const DefaultMemSize = 1024

const (
	OpHALT = iota
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
)

type Word uint64

type Machine struct {
	Memory        []Word
	A, I, P, X, Y Word
}

func New() *Machine {
	return &Machine{
		Memory: make([]Word, DefaultMemSize),
	}
}

func (g *Machine) Run() error {
	for {
		fmt.Printf("P: %d NextOp: %d A: %d I: %d X: %d Y: %d\n", g.P, g.Memory[g.P], g.A, g.I, g.X, g.Y)
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

// Map of assembly commands to OP codes
var commands = map[string]Word{
	"noop": OpNOOP,
	"halt": OpHALT,
	"inca": OpINCA,
	"seta": OpSETA,
	"deca": OpDECA,
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func normalize(s []byte) []byte {
	rawToken := bytes.TrimSpace(s)
	rawToken = bytes.ToLower(s)
	return rawToken
}

// scanLine is borrowed from bufio/scan with an added token handling splitting on
// ; as well as new lines.
func scanLine(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, normalize(dropCR(data[0:i])), nil
	}
	if i := bytes.IndexByte(data, ';'); i >= 0 {
		return i + 1, normalize(data[0:i]), nil
	}
	if i := bytes.IndexByte(data, ' '); i > 0 {
		return i + 1, normalize(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), normalize(dropCR(data)), nil
	}
	// Request more data.
	return 0, nil, nil
}

type Token struct {
	Raw    string
	Opcode Word
}

func Tokenize(reader io.Reader) ([]Token, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(scanLine)
	var tokens []Token
	for scanner.Scan() {
		tk := scanner.Text()
		// Trim space and lowercase Token
		tokens = append(tokens, Token{Raw: tk})
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return tokens, nil
}

type assembler struct {
	tks  []Token
	isns []Word
	err  error
}

type assemblerState func(s *assembler) assemblerState

func wantInstruction(s *assembler) assemblerState {
	if len(s.tks) == 0 {
		return nil
	}
	t := s.tks[0]
	command, ok := commands[t.Raw]
	if !ok {
		s.err = fmt.Errorf("unknown command %q", t.Raw)
		return nil
	}
	s.tks = s.tks[1:]
	s.isns = append(s.isns, command)
	return wantInstruction
}

func Assemble(reader io.Reader) (program []Word, err error) {
	tokens, err := Tokenize(reader)
	if err != nil {
		return nil, err
	}
	s := &assembler{tks: tokens}
	for state := wantInstruction; state != nil; {
		state = state(s)
	}
	if s.err != nil {
		return nil, s.err
	}
	return s.isns, nil
}

func AssembleFromFile(filename string) ([]Word, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return Assemble(file)
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
