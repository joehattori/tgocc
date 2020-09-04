package main

import (
	"fmt"
	"os"

	"github.com/joehattori/tgocc/parser"
	"github.com/joehattori/tgocc/tokenizer"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: ./tccgo <filename>")
		return
	}
	t := tokenizer.NewTokenizer(os.Args[1], true)
	toks := t.Tokenize()
	parser := parser.NewParser(toks)
	parser.Parse()
	parser.Res.Gen()
}
