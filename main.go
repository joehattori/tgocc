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
	Program(tokens)
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")
	fmt.Println("main:")

	// extend stack for local variables (named a to z for now)
	fmt.Println("	push rbp")
	fmt.Println("	mov rbp, rsp")
	fmt.Printf("	sub rsp, %d\n", 8*28)

	for _, code := range Code {
		Gen(code)
		fmt.Println("	pop rax")
	}

	fmt.Println("	mov rsp, rbp")
	fmt.Println("	pop rbp")
	fmt.Println("	ret")
}
