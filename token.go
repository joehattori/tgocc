package main

import (
	"regexp"
	"strconv"
	"strings"
)

type tokenKind int

const (
	tkReserved = iota
	tkNum
	tkEOF
)

// Token is a type to describe token
type Token struct {
	kind tokenKind
	val  int
	str  string
}

// Tokenize returns the tokenized input
func Tokenize(s string) []*Token {
	var toks []*Token
	for {
		if s == "" {
			break
		}

		s = strings.TrimLeft(s, " \n\t")

		regNum := regexp.MustCompile(`^[0-9]+`)
		if regNum.MatchString(s) {
			numStr := regNum.FindString(s)
			num, _ := strconv.Atoi(numStr)
			curTok := &Token{kind: tkNum, val: num}
			toks = append(toks, curTok)
			s = regNum.ReplaceAllString(s, "")
			continue
		}

		regOp := regexp.MustCompile(`^[+-*/]`)
		if regOp.MatchString(s) {
			curTok := &Token{kind: tkReserved, str: s}
			toks = append(toks, curTok)
			s = regOp.ReplaceAllString(s, "")
			continue
		}

		panic("unexpected input")
	}
	toks = append(toks, &Token{kind: tkEOF, str: s})
	return toks
}
