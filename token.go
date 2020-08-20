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

// Tokenized represents tokenized input generated in token.go
type Tokenized struct {
	toks  []*Token
	curFn *FnNode
	res   *Ast
}

// Ast is an array of nodes which represents whole program input (technically not a tree, but let's call this Ast still.)
type Ast struct {
	fns   []*FnNode
	gvars []*GVar
}

const idRegexp string = `^[a-zA-Z_]+\w*`

// Tokenize returns the tokenized input
func Tokenize(s string) *Tokenized {
	var toks []*Token
	for {
		if s == "" {
			break
		}

		s = strings.TrimSpace(s)

		if r := regexp.MustCompile(`^\d+`); r.MatchString(s) {
			numStr := r.FindString(s)
			num, _ := strconv.Atoi(numStr)
			toks = append(toks, &Token{kind: tkNum, val: num, length: len(numStr)})
			s = s[len(numStr):]
			continue
		}

		if r := regexp.MustCompile(`^(if|else|while|for|return|int|char|sizeof)\W`); r.MatchString(s) {
			cStr := r.FindString(s)
			toks = append(toks, &Token{kind: tkReserved, str: s, length: len(cStr) - 1})
			s = s[len(cStr)-1:]
			continue
		}

		if hasMultipleCharactorOperator(s) {
			toks = append(toks, &Token{kind: tkReserved, str: s, length: 2})
			s = s[2:]
			continue
		}

		if strings.Contains("+-*/(){}[]<>;=,&", s[:1]) {
			toks = append(toks, &Token{kind: tkReserved, str: s, length: 1})
			s = s[1:]
			continue
		}

		if r := regexp.MustCompile(idRegexp); r.MatchString(s) {
			varStr := r.FindString(s)
			toks = append(toks, &Token{kind: tkID, str: s, length: len(varStr)})
			s = s[len(varStr):]
			continue
		}

		panic(fmt.Sprintf("unexpected input %s\n", s))
	}
	toks = append(toks, &Token{kind: tkEOF, str: s})
	return &Tokenized{toks: toks, res: &Ast{}}
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
