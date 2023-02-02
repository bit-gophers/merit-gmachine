package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
)

//go:embed compiler/main.go
var mainData []byte

func main() {
	if err := makeCompiler(); err != nil {
		log.Fatal(err)
	}
}

func makeCompiler() error {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.Remove(tmpDir)
	mainFile := tmpDir + "/main.go"
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
