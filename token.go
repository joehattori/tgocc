package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type tokenKind int

const (
	tkReserved = iota
	tkNum
	tkID
	tkEOF
)

// Token is a type to describe token
type Token struct {
	kind   tokenKind
	val    int
	str    string
	length int
}

// Tokenize returns the tokenized input
func Tokenize(s string) []*Token {
	var toks []*Token
	for {
		if s == "" {
			break
		}

		s = strings.TrimLeft(s, " \t")

		regRet := regexp.MustCompile(`^return[^\w_]`)
		if regRet.MatchString(s) {
			toks = append(toks, &Token{kind: tkReserved, str: s, length: len("return")})
			s = s[len("return"):]
			continue
		}

		regNum := regexp.MustCompile(`^[0-9]+`)
		if regNum.MatchString(s) {
			numStr := regNum.FindString(s)
			num, _ := strconv.Atoi(numStr)
			toks = append(toks, &Token{kind: tkNum, val: num, length: len(numStr)})
			s = regNum.ReplaceAllString(s, "")
			continue
		}

		if hasMultipleCharactorOperator(s) {
			toks = append(toks, &Token{kind: tkReserved, str: s, length: 2})
			s = s[2:]
			continue
		}

		regSingleCharOp := regexp.MustCompile(`^[\+\-\*/\(\)<>;=]`)
		if regSingleCharOp.MatchString(s) {
			toks = append(toks, &Token{kind: tkReserved, str: s, length: 1})
			s = s[1:]
			continue
		}

		regVar := regexp.MustCompile(`^[a-zA-Z_]+[\w_]*`)
		if regVar.MatchString(s) {
			varStr := regVar.FindString(s)
			toks = append(toks, &Token{kind: tkID, str: s, length: len(varStr)})
			s = s[len(varStr):]
			continue
		}

		panic(fmt.Sprintf("unexpected input %s\n", s))
	}
	toks = append(toks, &Token{kind: tkEOF, str: s})
	return toks
}

func hasMultipleCharactorOperator(s string) bool {
	ops := []string{"==", "!=", "<=", ">="}
	for _, op := range ops {
		if strings.HasPrefix(s, op) {
			return true
		}
	}
	return false
}

//LVars represents the array of local variables
var LVars []*LVar

// LVar represents local variable. `offset` is the offset from rbp
type LVar struct {
	name   string
	offset int
}
