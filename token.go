package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var idRegexp = regexp.MustCompile(`^[a-zA-Z_]+\w*`)
var typeRegexp = regexp.MustCompile(`^(int|char|long|short|struct)\W`)

type token interface {
	getStr() string
}

type eofTok struct{}

type idTok struct {
	id  string
	len int
}

type numTok struct {
	val int
	len int
}

type reservedTok struct {
	str string
	len int
}

type strTok struct {
	content string
	len     int
}

func (e *eofTok) getStr() string      { return "" }
func (i *idTok) getStr() string       { return i.id }
func (n *numTok) getStr() string      { return "" }
func (r *reservedTok) getStr() string { return r.str }
func (s *strTok) getStr() string      { return "" }

func newEOFTok() *eofTok                            { return &eofTok{} }
func newIDTok(str string, l int) *idTok             { return &idTok{str, l} }
func newNumTok(val int, l int) *numTok              { return &numTok{val, l} }
func newReservedTok(str string, l int) *reservedTok { return &reservedTok{str, l} }
func newStrTok(content string, l int) *strTok       { return &strTok{content, l} }

type tokenized struct {
	curFnName string
	curScope  *scope
	res       *ast
	toks      []token
}

func newTokenized(toks []token) *tokenized {
	return &tokenized{curScope: &scope{}, res: &ast{}, toks: toks}
}

func (t *tokenized) spawnScope() {
	t.curScope = newScope(t.curScope)
}

func (t *tokenized) rewindScope() {
	offset := t.curScope.curOffset
	base := t.curScope.baseOffset
	for _, v := range t.curScope.vars {
		offset += v.getType().size()
		if lv, ok := v.(*lVar); ok {
			lv.offset = offset + base
		}
	}
	t.curScope.curOffset += offset
	t.curScope = t.curScope.super
	// maybe this is not necessary if we zero out the memory for child scope?
	t.curScope.curOffset += offset
}

func (t *tokenized) findVar(s string) variable {
	v := t.searchVar(s)
	if v == nil {
		log.Fatalf("undefined variable %s", s)
	}
	return v
}

func (t *tokenized) searchStructTag(tag string) *structTag {
	scope := t.curScope
	for scope != nil {
		if tag := scope.searchStructTag(tag); tag != nil {
			return tag
		}
		scope = scope.super
	}
	return nil
}

func (t *tokenized) searchVar(varName string) variable {
	scope := t.curScope
	for scope != nil {
		if v := scope.searchVar(varName); v != nil {
			return v
		}
		scope = scope.super
	}
	return nil
}

type ast struct {
	fns   []*fnNode
	gVars []*gVar
}

type tokenizer struct {
	input string
	pos   int
	res   tokenized
}

func newTokenizer() *tokenizer {
	return new(tokenizer)
}

func (t *tokenizer) cur() string {
	return t.input[t.pos:]
}

func (t *tokenizer) head() rune {
	r, _ := utf8.DecodeRuneInString(t.cur())
	return r
}

func (t *tokenizer) isComment() bool {
	if strings.HasPrefix(t.cur(), "//") {
		t.pos += 2
		for t.head() != '\n' {
			t.pos++
		}
		t.pos++
		return true
	}
	if strings.HasPrefix(t.cur(), "/*") {
		t.pos += 2
		for !strings.HasPrefix(t.cur(), "*/") {
			if t.cur() == "" {
				log.Fatalf("comment unclosed: %s", t.input)
			}
			t.pos++
		}
		t.pos += 2
		return true
	}
	return false
}

func (t *tokenizer) readCharLiteral() token {
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
	return newNumTok(c, 1)
}

func (t *tokenizer) readDigit() token {
	s := t.cur()
	r := regexp.MustCompile(`^\d+`)
	if !r.MatchString(s) {
		return nil
	}
	numStr := r.FindString(s)
	num, _ := strconv.Atoi(numStr)
	t.pos += len(numStr)
	return newNumTok(num, len(numStr))
}

func (t *tokenizer) readID() token {
	s := t.cur()
	if !idRegexp.MatchString(s) {
		return nil
	}
	l := len(idRegexp.FindString(s))
	t.pos += l
	return newIDTok(s, l)
}

func (t *tokenizer) readMultiCharOp() token {
	ops := []string{"==", "!=", "<=", ">=", "->"}
	s := t.cur()
	for _, op := range ops {
		if strings.HasPrefix(s, op) {
			t.pos += utf8.RuneCountInString(op)
			return newReservedTok(s, len(op))
		}
	}
	return nil
}

func (t *tokenizer) readReserved() token {
	s := t.cur()
	r := regexp.MustCompile(`^(if|else|while|for|return|sizeof|typedef)\W`)
	if r.MatchString(s) {
		l := len(r.FindString(s)) - 1
		t.pos += l
		return newReservedTok(s, l)
	}
	if typeRegexp.MatchString(s) {
		l := len(typeRegexp.FindString(s)) - 1
		t.pos += l
		return newReservedTok(s, l)
	}
	return nil
}

func (t *tokenizer) readRuneFrom(s string) token {
	if !strings.ContainsRune(s, t.head()) {
		return nil
	}
	cur := t.cur()
	t.pos++
	return newReservedTok(cur, 1)
}

func (t *tokenizer) readStrLiteral() token {
	if t.head() != '"' {
		return nil
	}
	t.pos++
	var s string
	// TODO: escape charator for others. e.g) \t
	for t.head() != '"' {
		if t.head() == '\\' {
			s += string(t.head())
			t.pos++
		}
		s += string(t.head())
		t.pos++
	}
	t.pos++
	return newStrTok(s, len(s))
}

func (t *tokenizer) trimSpace() {
	for unicode.IsSpace(t.head()) {
		t.pos++
	}
}

func (t *tokenizer) tokenize(input string) *tokenized {
	t.input = input
	var toks []token
	for {
		t.trimSpace()
		if t.isComment() {
			continue
		}
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

		if tok := t.readRuneFrom("+-*/(){}[]<>;=,&."); tok != nil {
			toks = append(toks, tok)
			continue
		}

		if tok := t.readID(); tok != nil {
			toks = append(toks, tok)
			continue
		}

		log.Fatalf("unexpected input %s\n", s)
	}
	toks = append(toks, newEOFTok())
	return newTokenized(toks)
}
