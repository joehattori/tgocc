package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: ./tccgo <filename>")
		return
	}
	t := newTokenizer(os.Args[1], true)
	toks := t.tokenize()
	parser := newParser(toks)
	parser.parse()
	parser.res.gen()
}
