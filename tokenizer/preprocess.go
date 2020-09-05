package tokenizer

import (
	"log"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var macros = map[string]macro{}

type macro interface {
	aMacro() // dummy method to avoid type errors
}

type fnMacro struct {
	params []Token
	body   []Token
}

func (*fnMacro) aMacro() {}

func newFnMacro(params []Token, body []Token) *fnMacro {
	return &fnMacro{params, body}
}

type objMacro []Token

func (*objMacro) aMacro() {}

type preprocessor struct {
	toks     []Token
	addEOF   bool
	filePath string
}

func newPreprocessor(toks []Token, addEOF bool, filePath string) *preprocessor {
	return &preprocessor{toks, addEOF, filePath}
}

func (p *preprocessor) beginsWith(s string) bool {
	return strings.HasPrefix(p.toks[0].Str(), s)
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

func (p *preprocessor) expect(str string) {
	if r, ok := p.toks[0].(*ReservedTok); ok &&
		r.Len() == utf8.RuneCountInString(str) && strings.HasPrefix(r.Str(), str) {
		p.popToks()
		return
	}
	log.Fatalf("%s was expected but got %s", str, p.toks[0].Str())
}

func (p *preprocessor) expectID() (tok *IDTok) {
	tok, _ = p.toks[0].(*IDTok)
	if tok == nil {
		log.Fatalf("Id was expected but got %s", p.toks[0].Str())
	}
	p.popToks()
	return
}

func (p *preprocessor) consumeStr() (tok *StrTok, ok bool) {
	defer func() {
		if ok {
			p.popToks()
		}
	}()
	tok, ok = p.toks[0].(*StrTok)
	return
}

func (p *preprocessor) popToks() {
	p.toks = p.toks[1:]
}

func (p *preprocessor) readParam() (param []Token) {
	level := 0
	for !p.isEOF() {
		if level == 0 {
			if p.beginsWith(")") || p.beginsWith(",") {
				return
			}
		}
		cur := p.toks[0]
		if p.beginsWith("(") {
			level++
		} else if p.beginsWith(")") {
			level--
		}
		param = append(param, cur)
		p.popToks()
	}
	return
}

func (p *preprocessor) readParams() (params [][]Token) {
	p.expect("(")
	if p.consume(")") {
		return
	}
	params = append(params, p.readParam())
	for !p.consume(")") {
		p.expect(",")
		params = append(params, p.readParam())
	}
	return
}

func (p *preprocessor) Preprocess() []Token {
	var output []Token
	for !p.isEOF() {
		if p.consume("\n") {
			continue
		}
		cur := p.toks[0]
		if id, ok := p.consumeID(); ok {
			if m, ok := macros[id.Str()]; ok {
				switch m := m.(type) {
				case *fnMacro:
					params := p.readParams()
					if len(params) != len(m.params) {
						log.Fatalf("Number of parameters of macro %s does not match", id.Str())
					}
					for _, tok := range m.body {
						if p, ok := tok.(*paramTok); ok {
							output = append(output, params[p.idx]...)
						} else {
							output = append(output, tok)
						}
					}
				case *objMacro:
					output = append(output, *m...)
				}
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
			id := p.expectID()
			macros[id.Str()] = p.define()
		} else if p.consume("include") {
			newTok := NewTokenizer(p.includePath(), false)
			output = append(output, newTok.Tokenize()...)
		}
	}
	if p.addEOF {
		output = append(output, newEOFTok())
	}
	return output
}

func (p *preprocessor) readUntilEOL() (toks []Token) {
	for !p.consume("\n") {
		toks = append(toks, p.toks[0])
		p.popToks()
	}
	return
}

func (p *preprocessor) includePath() string {
	if strTok, isStr := p.consumeStr(); isStr {
		relPath := strTok.Str()
		return filepath.Join(filepath.Dir(p.filePath), strings.TrimRight(relPath, string('\000')))
	}
	// TODO
	//p.expect("<")
	//var filePath string
	//for !p.consume(">") {
	//filePath += p.toks[0].Str()
	//p.popToks()
	//}
	//return path.Join("/usr/include", filePath)
	return ""
}

func (p *preprocessor) define() macro {
	if p.beginsWith("(") {
		return p.defineFnLike()
	}
	return p.defineObjLike()
}

func (p *preprocessor) defineFnLike() *fnMacro {
	p.consume("(")
	var params []Token
	params = append(params, p.expectID())

	for !p.consume(")") {
		p.expect(",")
		params = append(params, p.expectID())
	}

	getIndexInParams := func(name string) int {
		for i, param := range params {
			if param.Str() == name {
				return i
			}
		}
		return -1
	}

	var body []Token
	for _, tok := range p.readUntilEOL() {
		if idx := getIndexInParams(tok.Str()); idx >= 0 {
			body = append(body, newParamTok(idx))
		} else {
			body = append(body, tok)
		}
	}
	return newFnMacro(params, body)
}

func (p *preprocessor) defineObjLike() *objMacro {
	ret := new(objMacro)
	*ret = p.readUntilEOL()
	return ret
}
