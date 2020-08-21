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
	t := newTokenizer()
	toks := t.tokenizeInput(os.Args[1])
	toks.parse()
	toks.res.gen()
}
