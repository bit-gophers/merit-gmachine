package gmachine

import (
	"bytes"
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

func Tokenize(data string) ([]Token, error) {
	return NewTokenizer().Run(data)
}

func NewTokenizer() *tokenizer {
	t := new(tokenizer)
	t.line = 1
	t.Log = new(bytes.Buffer)
	return t
}

type tokenizer struct {
	input            []rune
	Log              *bytes.Buffer
	start, pos, line int
	result           []Token
	err              error
}

func (t *tokenizer) Run(data string) ([]Token, error) {
	t.input = []rune(data)
	for state := wantToken; state != nil; {
		state = state(t)
		if t.err != nil {
			return nil, t.err
		}
	}
	return t.result, nil
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
	fmt.Fprintln(t.Log, args...)
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
		case '\'':
			return inRuneLiteral
		case eof:
			t.emit()
			return nil
		}
	}
}

func inRuneLiteral(t *tokenizer) stateFunc {
	for {
		t.logState("inRuneLiteral")
		switch t.next() {
		case '\'':
			t.emit()
			return wantToken
		case eof:
			t.err = fmt.Errorf("unexpected EOF in rune literal")
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
