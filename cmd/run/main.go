package main

import (
	gmachine "github.com/bit-gophers/merit-gmachine"
	"log"
	"os"
)

func main() {
	g := gmachine.New()
	err := g.AssembleAndRunFromFile(os.Args[1], true)
	if err != nil {
		log.Fatalln(err)
	}
}
