package main

import (
	"log"
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
	tkStr
)

var idRegexp = regexp.MustCompile(`^[a-zA-Z_]+\w*`)

// TODO: make token an interface

type token struct {
	kind    tokenKind
	val     int
	content string
	str     string
	len     int
}

func newNumtoken(val int, l int) *token {
	return &token{kind: tkNum, val: val, len: l}
}

func newStrtoken(content string, l int) *token {
	return &token{kind: tkStr, content: content, len: l}
}

func newToken(kind tokenKind, str string, l int) *token {
	return &token{kind: kind, str: str, len: l}
}

type tokenized struct {
	toks  []*token
	curFn *fnNode
	res   *ast
}

func newTokenized(toks []*token) *tokenized {
	return &tokenized{toks: toks, res: &ast{}}
}

type ast struct {
	fns   []*fnNode
	gvars []*gVar
}

type tokenizer struct {
	input string
	pos   int
	res   tokenized
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

func (t *tokenizer) readCharLiteral() *token {
	if t.head() != '\'' {
		return nil
	}
	t.pos++
	c := int(t.head())
	t.pos++
	if t.head() != '\'' {
		log.Fatalf("Char literal is too long: %s", t.input[t.pos:])
	}
	t.pos++
	return newNumtoken(c, 1)
}

func (t *tokenizer) readDigit() *token {
	s := t.cur()
	r := regexp.MustCompile(`^\d+`)
	if !r.MatchString(s) {
		return nil
	}
	numStr := r.FindString(s)
	num, _ := strconv.Atoi(numStr)
	t.pos += len(numStr)
	return newNumtoken(num, len(numStr))
}

func (t *tokenizer) readID() *token {
	s := t.cur()
	if !idRegexp.MatchString(s) {
		return nil
	}
	l := len(idRegexp.FindString(s))
	t.pos += l
	return newToken(tkID, s, l)
}

func (t *tokenizer) readMultiCharOp() *token {
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

func (t *tokenizer) readReserved() *token {
	s := t.cur()
	r := regexp.MustCompile(`^(if|else|while|for|return|int|char|sizeof)\W`)
	if !r.MatchString(s) {
		return nil
	}
	l := len(r.FindString(s)) - 1
	t.pos += l
	return newToken(tkReserved, s, l)
}

func (t *tokenizer) readRuneFrom(s string) *token {
	if !strings.ContainsRune(s, t.head()) {
		return nil
	}
	cur := t.cur()
	t.pos++
	return newToken(tkReserved, cur, 1)
}

func (t *tokenizer) readStrLiteral() *token {
	if t.head() != '"' {
		return nil
	}
	t.pos++
	var s string
	// TODO: escape charator
	for t.head() != '"' {
		s += string(t.head())
		t.pos++
	}
	t.pos++
	return newStrtoken(s, len(s))
}

func (t *tokenizer) tokenizeInput(input string) *tokenized {
	t.input = input
	var toks []*token
	for {
		t.trimSpace()
		s := t.cur()
		if s == "" {
			break
		}

		if tok := t.readStrLiteral(); tok != nil {
			toks = append(toks, tok)
			continue
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

		log.Fatalf("unexpected input %s\n", s)
	}
	toks = append(toks, newToken(tkEOF, t.cur(), 0))
	return newTokenized(toks)
}
