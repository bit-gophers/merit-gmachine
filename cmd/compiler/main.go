package main

import (
	_ "embed"
	"fmt"
	"os"

	gmachine "github.com/bit-gophers/merit-gmachine"
)

//go:embed target.g
var mainData string

func main() {
	g := gmachine.New()
	err := g.AssembleAndRunFromString(mainData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
