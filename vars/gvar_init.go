package vars

import (
	"fmt"
	"log"
	"strings"

	"github.com/joehattori/tgocc/types"
)

type (
	// GVarInit is the interface of global variable initializer.
	GVarInit interface {
		Gen(types.Type)
	}

	// GVarInitArr represents an array initializer of global variable.
	GVarInitArr struct {
		body []GVarInit
	}

	// GVarInitLabel represents a reference to another global variable
	GVarInitLabel struct {
		label string
	}

	// GVarInitStr represents a raw string global initializer.
	GVarInitStr struct {
		Content string
	}

	// GVarInitInt represents a integer initializer.
	GVarInitInt struct {
		val int64
		sz  int
	}

	// GVarInitZero is used to zero out the given length on global variable initialization.
	GVarInitZero struct {
		len int
	}
)

func NewGVarInitArr(body []GVarInit) *GVarInitArr {
	return &GVarInitArr{body}
}

func NewGVarInitLabel(label string) *GVarInitLabel {
	return &GVarInitLabel{label}
}

func NewGVarInitStr(content string) *GVarInitStr {
	return &GVarInitStr{content}
}

func NewGVarInitInt(i int64, sz int) *GVarInitInt {
	return &GVarInitInt{i, sz}
}

func NewGVarInitZero(len int) *GVarInitZero {
	return &GVarInitZero{len}
}

func (init *GVarInitArr) Gen(t types.Type) {
	for i, e := range init.body {
		switch t := t.(type) {
		case *types.Arr:
			e.Gen(t.Base())
		case *types.Struct:
			e.Gen(t.Members[i].Type)
		default:
			e.Gen(t)
		}
	}
}

func (init *GVarInitLabel) Gen(_ types.Type) {
	fmt.Printf("	.quad %s\n", init.label)
}

func (init *GVarInitStr) Gen(_ types.Type) {
	trimmed := strings.TrimRight(init.Content, string('\000'))
	fmt.Printf("	.string \"%s\"\n", trimmed)
	fmt.Printf("	.zero %d\n", len(init.Content)-len(trimmed))
}

func (init *GVarInitInt) Gen(_ types.Type) {
	switch init.sz {
	case 1:
		fmt.Printf("	.byte %d\n", init.val)
	case 2:
		fmt.Printf("	.value %d\n", init.val)
	case 4:
		fmt.Printf("	.long %d\n", init.val)
	case 8:
		fmt.Printf("	.quad %d\n", init.val)
	default:
		log.Fatalf("Unhandled type size %d on global variable initialization.", init.sz)
	}
}

func (init *GVarInitZero) Gen(_ types.Type) {
	fmt.Printf("	.zero %d\n", init.len)
}
