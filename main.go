package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: tccgo <program>")
		return
	}
	tokens := Tokenize(os.Args[1])
	_, node := Expr(tokens)
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")
	fmt.Println("main:")
	Gen(node)
	fmt.Println("	pop rax")
	fmt.Println("	ret")
}
