package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

//go:embed compiler/main.go
var mainData []byte

func main() {
	var fn string
	if len(os.Args) < 2 {
		fmt.Println("Filename is required")
		os.Exit(1)
	}
	fn = os.Args[1]
	if err := makeCompiler(fn); err != nil {
		log.Fatal(err)
	}
}

func makeCompiler(fn string) error {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.Remove(tmpDir)

	mainFile := tmpDir + "/main.go"
	file, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer file.Close()

	runFile := tmpDir + "/target.g"
	openedFile, err := os.OpenFile(runFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer openedFile.Close()

	_, err = io.Copy(openedFile, file)
	if err != nil {
		return err
	}
	if err := os.WriteFile(mainFile, mainData, 0644); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", "./gmachine", mainFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
		return err
	}
	return nil
}
