package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: ./tccgo <filename>")
		return
	}
	t := newTokenizer()
	input, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	parser := t.tokenize(string(input))
	parser.parse()
	parser.res.gen()
}
