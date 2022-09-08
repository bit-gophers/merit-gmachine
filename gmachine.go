// Package gmachine implements a simple virtual CPU, known as the G-machine.
package gmachine

import (
	"fmt"
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

func (g *Machine) AssembleAndRun(program string) error {
	commands, err := Assemble(program)
	if err != nil {
		return err
	}
	return g.RunProgram(commands)
}

// Map of assembly commands to OP codes
var commands = map[string]Word{
	"noop": OpNOOP,
	"halt": OpHALT,
}

func Assemble(s string) (program []Word, err error) {
	// Normalize input string
	s = strings.ToLower(s)
	// Split into tokens
	tokens := strings.Split(s, ";")
	// Iterate over tokens
	for _, token := range tokens {
		// Trim space around token
		token = strings.TrimSpace(token)
		// Convert each token to op commands
		command, ok := commands[token]
		if !ok {
			return nil, fmt.Errorf("unknown command %q", token)
		}
		program = append(program, command)
	}
	return program, nil
}
