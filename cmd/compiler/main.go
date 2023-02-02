package main

import (
	"fmt"
	gmachine "github.com/bit-gophers/merit-gmachine"
	"os"
)

func main() {
	g := gmachine.New()
	err := g.AssembleAndRunFromFile(fn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
