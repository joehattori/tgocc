package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	idRegexp   = regexp.MustCompile(`^[a-zA-Z_]+\w*`)
	typeRegexp = regexp.MustCompile(
		`^(int|char|long|short|struct|void|_Bool|typedef|enum|static|continue|switch|case|default|extern|do|define)\W`)
)

type token interface {
	getStr() string
}

type eofTok struct{}

type idTok struct {
	id  string
	len int
}

type numTok struct {
	val int64
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
func newNumTok(val int64, l int) *numTok            { return &numTok{val, l} }
func newReservedTok(str string, l int) *reservedTok { return &reservedTok{str, l} }
func newStrTok(content string, l int) *strTok       { return &strTok{content, l} }

type ast struct {
	fns   []*fnNode
	gVars []*gVar
}

type tokenizer struct {
	input string
	pos   int
	res   parser
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
	c := int64(t.head())
	t.pos++
	if t.head() != '\'' {
		log.Fatalf("char literal is too long: %s", t.input[t.pos:])
	}
	t.pos++
	return newNumTok(c, 1)
}

func (t *tokenizer) readDigitLiteral() token {
	s := t.cur()
	r := regexp.MustCompile(`^(0(x|X)[[:xdigit:]]+|0(o|O)\d+|0(b|B)(0|1)+|\d+)`)
	if !r.MatchString(s) {
		return nil
	}
	numStr := r.FindString(s)
	numLen := utf8.RuneCountInString(numStr)
	num, err := strconv.ParseInt(numStr, 0, 64)
	if err != nil {
		log.Fatalf("invalid number literal: %s", s)
	}
	t.pos += numLen
	return newNumTok(num, numLen)
}

func (t *tokenizer) readID() token {
	s := t.cur()
	if !idRegexp.MatchString(s) {
		return nil
	}
	l := utf8.RuneCountInString(idRegexp.FindString(s))
	t.pos += l
	return newIDTok(s, l)
}

func (t *tokenizer) readMultiCharOp() token {
	ops := [...]string{
		"==", "!=", "<=", ">=", "->", "++", "--",
		"+=", "-=", "*=", "/=", "&&", "||",
		"<<=", ">>=", "<<", ">>",
	}
	s := t.cur()
	for _, op := range ops {
		if strings.HasPrefix(s, op) {
			t.pos += utf8.RuneCountInString(op)
			return newReservedTok(s, utf8.RuneCountInString(op))
		}
	}
	return nil
}

func (t *tokenizer) readNewLine() token {
	if t.head() != '\n' {
		return nil
	}
	t.pos++
	return newReservedTok("\n", 1)
}

func (t *tokenizer) readReserved() token {
	s := t.cur()
	r := regexp.MustCompile(`^(if|else|while|for|return|sizeof|break)\W`)
	if r.MatchString(s) {
		l := utf8.RuneCountInString(r.FindString(s)) - 1
		t.pos += l
		return newReservedTok(s, l)
	}
	if typeRegexp.MatchString(s) {
		l := utf8.RuneCountInString(typeRegexp.FindString(s)) - 1
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
	s += string('\000')
	t.pos++
	return newStrTok(s, len(s))
}

func (t *tokenizer) trimSpace() {
	for unicode.IsSpace(t.head()) {
		t.pos++
	}
}

func (t *tokenizer) tokenize(input string) *parser {
	t.input = input
	var toks []token
	for {
		// new line will be omitted in preprocessor, but still needed to parse #include ... and #define ...
		if tok := t.readNewLine(); tok != nil {
			toks = append(toks, tok)
			continue
		}

		t.trimSpace()

		s := t.cur()
		if s == "" {
			break
		}
		if t.isComment() {
			continue
		}

		if tok := t.readStrLiteral(); tok != nil {
			toks = append(toks, tok)
			continue
		}

		if tok := t.readCharLiteral(); tok != nil {
			toks = append(toks, tok)
			continue
		}

		if tok := t.readDigitLiteral(); tok != nil {
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

		if tok := t.readRuneFrom("+-*/(){}[]<>;=,&.!|^:?~#"); tok != nil {
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
	parser := newParser(toks)
	parser.preprocess()
	return parser
}
