package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokenKind int

const (
	tkReserved = iota
	tkNum
	tkID
	tkEOF
)

const idRegexp string = `^[a-zA-Z_]+\w*`

// Token is a type to describe token
type Token struct {
	kind tokenKind
	val  int
	str  string
	len  int
}

func newNumToken(val int, l int) *Token {
	return &Token{kind: tkNum, val: val, len: l}
}

func newToken(kind tokenKind, str string, l int) *Token {
	return &Token{kind: kind, str: str, len: l}
}

// Tokenized represents tokenized input generated in token.go
type Tokenized struct {
	toks  []*Token
	curFn *FnNode
	res   *Ast
}

func newTokenized(toks []*Token) *Tokenized {
	return &Tokenized{toks: toks, res: &Ast{}}
}

// Ast is an array of nodes which represents whole program input (technically not a tree, but let's call this Ast still.)
type Ast struct {
	fns   []*FnNode
	gvars []*GVar
}

type tokenizer struct {
	input string
	pos   int
	res   Tokenized
}

func newTokenizer() *tokenizer {
	return new(tokenizer)
}

func (t *tokenizer) head() rune {
	r, _ := utf8.DecodeRuneInString(t.cur())
	return r
}

func (t *tokenizer) cur() string {
	return t.input[t.pos:]
}

func (t *tokenizer) trimSpace() {
	for unicode.IsSpace(t.head()) {
		t.pos++
	}
}

func (t *tokenizer) isAny(s string) bool {
	return strings.ContainsRune(s, t.head())
}

func (t *tokenizer) readCharLiteral() *Token {
	if t.head() != '\'' {
		return nil
	}
	t.pos++
	c := int(t.head())
	t.pos++
	if t.head() != '\'' {
		panic(fmt.Sprintf("Char literal is too long: %s", t.input[t.pos:]))
	}
	t.pos++
	return &Token{kind: tkNum, val: c, len: 1}
}

func (t *tokenizer) readDigit() *Token {
	s := t.cur()
	r := regexp.MustCompile(`^\d+`)
	if !r.MatchString(s) {
		return nil
	}
	numStr := r.FindString(s)
	num, _ := strconv.Atoi(numStr)
	t.pos += len(numStr)
	return newNumToken(num, len(numStr))
}

func (t *tokenizer) readID() *Token {
	s := t.cur()
	r := regexp.MustCompile(idRegexp)
	if !r.MatchString(s) {
		return nil
	}
	l := len(r.FindString(s))
	t.pos += l
	return newToken(tkID, s, l)
}

func (t *tokenizer) readMultiCharOp() *Token {
	ops := []string{"==", "!=", "<=", ">="}
	s := t.cur()
	for _, op := range ops {
		if strings.HasPrefix(s, op) {
			t.pos += 2
			return newToken(tkReserved, s, len(op))
		}
	}
	return nil
}

func (t *tokenizer) readReserved() *Token {
	s := t.cur()
	r := regexp.MustCompile(`^(if|else|while|for|return|int|char|sizeof)\W`)
	if !r.MatchString(s) {
		return nil
	}
	l := len(r.FindString(s)) - 1
	t.pos += l
	return newToken(tkReserved, s, l)
}

func (t *tokenizer) readRuneFrom(s string) *Token {
	if !strings.ContainsRune(s, t.head()) {
		return nil
	}
	cur := t.cur()
	t.pos++
	return newToken(tkReserved, cur, 1)
}

func (t *tokenizer) tokenizeInput(input string) *Tokenized {
	t.input = input
	var toks []*Token
	for {
		t.trimSpace()
		s := t.cur()
		if s == "" {
			break
		}

		if tok := t.readCharLiteral(); tok != nil {
			toks = append(toks, tok)
			continue
		}

		if tok := t.readDigit(); tok != nil {
			toks = append(toks, tok)
			continue
		}

		if tok := t.readReserved(); tok != nil {
			toks = append(toks, tok)
			continue
		}

		if tok := t.readMultiCharOp(); tok != nil {
			toks = append(toks, tok)
			continue
		}

		if tok := t.readRuneFrom("+-*/(){}[]<>;=,&"); tok != nil {
			toks = append(toks, tok)
			continue
		}

		if tok := t.readID(); tok != nil {
			toks = append(toks, tok)
			continue
		}

		panic(fmt.Sprintf("unexpected input %s\n", s))
	}
	toks = append(toks, newToken(tkEOF, t.cur(), 0))
	return newTokenized(toks)
}
