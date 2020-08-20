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
	t := Tokenize(os.Args[1])
	t.parse()
	t.res.gen()
}
