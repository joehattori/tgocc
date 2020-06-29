package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: ./tccgo <program>")
		return
	}
	tokens := Tokenize(os.Args[1])
	Program(tokens)
	Gen()
}
