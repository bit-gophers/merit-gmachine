package gmachine

import (
	"fmt"
	"io"
	"os"
)

func Assemble(input io.Reader) (program []Word, err error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	tokens, err := Tokenize(string(data))
	if err != nil {
		return nil, err
	}
	argRequired := false
	for _, token := range tokens {
		switch token.Kind {
		case TokenComment:
			continue
		case TokenInstruction:
			if argRequired {
				return nil, fmt.Errorf("line %d: unexpected instruction %q", token.Line, token.RawToken)
			}
			argRequired = OpCode(token.Value).RequiresArgument()
		case TokenRuneLiteral, TokenNumberLiteral:
			argRequired = false
		default:
			return nil, fmt.Errorf("line %d: unknown token kine %q", token.Line, token.Kind)
		}
		program = append(program, token.Value)
	}
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
