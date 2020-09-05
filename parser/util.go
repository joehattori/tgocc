package parser

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/joehattori/tgocc/tokenizer"
	"github.com/joehattori/tgocc/vars"
)

func (p *Parser) beginsWith(s string) bool {
	return strings.HasPrefix(p.Toks[0].Str(), s)
}

func (p *Parser) consume(str string) bool {
	if r, ok := p.Toks[0].(*tokenizer.ReservedTok); ok &&
		r.Len() == utf8.RuneCountInString(str) &&
		strings.HasPrefix(r.Str(), str) {
		p.popToks()
		return true
	}
	return false
}

func (p *Parser) consumeID() (tok *tokenizer.IDTok, ok bool) {
	defer func() {
		if ok {
			p.popToks()
		}
	}()
	tok, ok = p.Toks[0].(*tokenizer.IDTok)
	return
}

func (p *Parser) consumeStr() (tok *tokenizer.StrTok, ok bool) {
	defer func() {
		if ok {
			p.popToks()
		}
	}()
	tok, ok = p.Toks[0].(*tokenizer.StrTok)
	return
}

func (p *Parser) isEOF() (ok bool) {
	_, ok = p.Toks[0].(*tokenizer.EOFTok)
	return
}

func (p *Parser) expect(str string) {
	if r, ok := p.Toks[0].(*tokenizer.ReservedTok); ok &&
		r.Len() == utf8.RuneCountInString(str) && strings.HasPrefix(r.Str(), str) {
		p.popToks()
		return
	}
	log.Fatalf("%s was expected but got %s", str, p.Toks[0].Str())
}

func (p *Parser) expectID() (tok *tokenizer.IDTok) {
	tok, _ = p.Toks[0].(*tokenizer.IDTok)
	if tok == nil {
		log.Fatalf("Id was expected but got %s", p.Toks[0].Str())
	}
	p.popToks()
	return
}

func (p *Parser) expectNum() (tok *tokenizer.NumTok) {
	tok, _ = p.Toks[0].(*tokenizer.NumTok)
	if tok == nil {
		log.Fatalf("Number was expected but got %s", p.Toks[0].Str())
	}
	p.popToks()
	return
}

func (p *Parser) expectStr() (tok *tokenizer.StrTok) {
	tok, _ = p.Toks[0].(*tokenizer.StrTok)
	if tok == nil {
		log.Fatalf("String literal was expected but got %s", p.Toks[0].Str())
	}
	p.popToks()
	return
}

func (p *Parser) isFunction() bool {
	orig := p.Toks
	defer func() { p.Toks = orig }()
	p.baseType()
	for p.consume("*") {
	}
	_, isID := p.consumeID()
	return isID && p.consume("(")
}

func (p *Parser) isType() (ret bool) {
	switch tok := p.Toks[0].(type) {
	case *tokenizer.IDTok:
		_, ret = p.searchVar(tok.Str()).(*vars.TypeDef)
	case *tokenizer.ReservedTok:
		ret = tok.IsType
	}
	return
}

func (p *Parser) popToks() {
	p.Toks = p.Toks[1:]
}

var gVarLabelCount int

func newGVarLabel() string {
	defer func() { gVarLabelCount++ }()
	return fmt.Sprintf(".L.data.%d", gVarLabelCount)
}
