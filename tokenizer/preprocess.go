package tokenizer

import (
	"log"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var macros = map[string][]Token{}

type preprocessor struct {
	toks     []Token
	addEOF   bool
	filePath string
}

func newPreprocessor(toks []Token, addEOF bool, filePath string) *preprocessor {
	return &preprocessor{toks, addEOF, filePath}
}

func (p *preprocessor) isEOF() (ok bool) {
	if len(p.toks) == 0 {
		return true
	}
	_, ok = p.toks[0].(*EOFTok)
	return
}

func (p *preprocessor) consume(str string) bool {
	if r, ok := p.toks[0].(*ReservedTok); ok &&
		r.len == utf8.RuneCountInString(str) &&
		strings.HasPrefix(r.str, str) {
		p.popToks()
		return true
	}
	return false
}

func (p *preprocessor) consumeID() (tok *IDTok, ok bool) {
	defer func() {
		if ok {
			p.popToks()
		}
	}()
	tok, ok = p.toks[0].(*IDTok)
	return
}

func (p *preprocessor) expectID() (tok *IDTok) {
	tok, _ = p.toks[0].(*IDTok)
	if tok == nil {
		log.Fatalf("Id was expected but got %s", p.toks[0].Str())
	}
	p.popToks()
	return
}

func (p *preprocessor) expectStr() (tok *StrTok) {
	tok, _ = p.toks[0].(*StrTok)
	if tok == nil {
		log.Fatalf("String literal was expected but got %s", p.toks[0].Str())
	}
	p.popToks()
	return
}

func (p *preprocessor) popToks() {
	p.toks = p.toks[1:]
}

func (p *preprocessor) Preprocess() []Token {
	var output []Token
	for !p.isEOF() {
		if p.consume("\n") {
			continue
		}
		cur := p.toks[0]
		if id, ok := p.consumeID(); ok {
			if macro, ok := macros[id.Str()]; ok {
				output = append(output, macro...)
			} else {
				output = append(output, cur)
			}
			continue
		}

		if !p.consume("#") {
			output = append(output, cur)
			p.popToks()
			continue
		}

		if p.consume("define") {
			var def []Token
			id := p.expectID()
			for !p.isEOF() {
				if p.consume("\n") {
					break
				}
				cur := p.toks[0]
				p.popToks()
				def = append(def, cur)
			}
			macros[id.Str()] = def
		} else if p.consume("include") {
			relPath := p.expectStr().Str()
			includePath := filepath.Join(filepath.Dir(p.filePath), strings.TrimRight(relPath, string('\000')))
			newTok := NewTokenizer(includePath, false)
			netoks := newTok.Tokenize()
			output = append(output, netoks...)
		}
	}
	if p.addEOF {
		output = append(output, newEOFTok())
	}
	return output
}
