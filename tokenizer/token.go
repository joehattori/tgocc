package tokenizer

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	idMatcher   = regexp.MustCompile(`^[a-zA-Z_]+\w*`)
	typeMatcher = regexp.MustCompile(
		`^(int|char|long|short|struct|void|_Bool|typedef|enum|static|extern|signed|unsigned|volatile)\W`)
)

type (
	Token interface {
		Str() string
		Len() int
	}

	EofTok struct{}

	IdTok struct {
		name string
		len  int
	}

	NumTok struct {
		Val int64
		len int
	}

	ReservedTok struct {
		str string
		len int
	}

	StrTok struct {
		content string
		len     int
	}
)

func (e *EofTok) Str() string      { return "" }
func (i *IdTok) Str() string       { return i.name }
func (n *NumTok) Str() string      { return fmt.Sprintf("%d", n.Val) }
func (r *ReservedTok) Str() string { return r.str }
func (s *StrTok) Str() string      { return s.content }

func (e *EofTok) Len() int      { return 0 }
func (i *IdTok) Len() int       { return i.len }
func (n *NumTok) Len() int      { return utf8.RuneCountInString(fmt.Sprintf("%d", n.Val)) }
func (r *ReservedTok) Len() int { return r.len }
func (s *StrTok) Len() int      { return s.len }

func NewEOFTok() *EofTok                            { return &EofTok{} }
func NewIDTok(str string, l int) *IdTok             { return &IdTok{str, l} }
func NewNumTok(val int64, l int) *NumTok            { return &NumTok{val, l} }
func NewReservedTok(str string, l int) *ReservedTok { return &ReservedTok{str, l} }
func NewStrTok(content string, l int) *StrTok       { return &StrTok{content, l} }

type Tokenizer struct {
	filePath string
	input    string
	addEOF   bool
	pos      int
	res      []Token
}

func NewTokenizer(path string, addEOF bool) *Tokenizer {
	return &Tokenizer{filePath: path, addEOF: addEOF}
}

func (t *Tokenizer) cur() string {
	return t.input[t.pos:]
}

func (t *Tokenizer) head() rune {
	r, _ := utf8.DecodeRuneInString(t.cur())
	return r
}

func (t *Tokenizer) isComment() bool {
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
				log.Fatalf("Comment unclosed: %s", t.input)
			}
			t.pos++
		}
		t.pos += 2
		return true
	}
	return false
}

func (t *Tokenizer) readCharLiteral() Token {
	if t.head() != '\'' {
		return nil
	}
	t.pos++
	c := int64(t.head())
	t.pos++
	if t.head() != '\'' {
		log.Fatalf("Char literal is too long: %s", t.input[t.pos:])
	}
	t.pos++
	return NewNumTok(c, 1)
}

func (t *Tokenizer) readDigitLiteral() Token {
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
	return NewNumTok(num, numLen)
}

func (t *Tokenizer) readID() Token {
	s := t.cur()
	if !idMatcher.MatchString(s) {
		return nil
	}
	id := idMatcher.FindString(s)
	l := utf8.RuneCountInString(id)
	t.pos += l
	return NewIDTok(id, l)
}

func (t *Tokenizer) readMultiCharOp() Token {
	ops := [...]string{
		"==", "!=", "<=", ">=", "->", "++", "--",
		"+=", "-=", "*=", "/=", "&&", "||",
		"<<=", ">>=", "<<", ">>",
	}
	s := t.cur()
	for _, op := range ops {
		if strings.HasPrefix(s, op) {
			t.pos += utf8.RuneCountInString(op)
			return NewReservedTok(op, utf8.RuneCountInString(op))
		}
	}
	return nil
}

func (t *Tokenizer) readNewLine() Token {
	if t.head() != '\n' {
		return nil
	}
	t.pos++
	return NewReservedTok("\n", 1)
}

func (t *Tokenizer) readReserved() Token {
	s := t.cur()
	r := regexp.MustCompile(`^(if|else|while|for|return|sizeof|break|continue|switch|case|default|do|define|include)\W`)
	if res := r.FindString(s); res != "" {
		l := utf8.RuneCountInString(res) - 1
		t.pos += l
		return NewReservedTok(res, l)
	}
	if res := typeMatcher.FindString(s); res != "" {
		l := utf8.RuneCountInString(res) - 1
		t.pos += l
		return NewReservedTok(res, l)
	}
	return nil
}

func (t *Tokenizer) readRuneFrom(s string) Token {
	if !strings.ContainsRune(s, t.head()) {
		return nil
	}
	cur := t.cur()
	t.pos++
	return NewReservedTok(cur[:1], 1)
}

func (t *Tokenizer) readStrLiteral() Token {
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
	return NewStrTok(s, len(s))
}

func (t *Tokenizer) trimSpace() {
	for unicode.IsSpace(t.head()) {
		t.pos++
	}
}

func (t *Tokenizer) Tokenize() []Token {
	input, err := ioutil.ReadFile(t.filePath)
	if err != nil {
		log.Fatal(err)
	}
	t.input = string(input)
	var toks []Token
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

		log.Fatalf("Unexpected input %s\n", s)
	}
	if t.addEOF {
		toks = append(toks, NewEOFTok())
	}
	p := &preprocessor{toks}
	return p.Preprocess(t.filePath, t.addEOF)
}
