package vars

import (
	"fmt"
	"log"
	"strings"

	"github.com/joehattori/tgocc/types"
)

type (
	GVarInit interface {
		gen(types.Type)
	}

	GVarInitArr struct {
		body []GVarInit
	}

	GVarInitLabel struct { // reference to another global var
		label string
	}

	GVarInitStr struct {
		Content string
	}

	GVarInitInt struct {
		val int64
		sz  int
	}

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

func (init *GVarInitArr) gen(t types.Type) {
	for i, e := range init.body {
		switch t := t.(type) {
		case *types.Arr:
			e.gen(t.Base())
		case *types.Struct:
			e.gen(t.Members[i].Type)
		default:
			e.gen(t)
		}
	}
}

func (init *GVarInitLabel) gen(_ types.Type) {
	fmt.Printf("	.quad %s\n", init.label)
}

func (init *GVarInitStr) gen(_ types.Type) {
	trimmed := strings.TrimRight(init.Content, string('\000'))
	fmt.Printf("	.string \"%s\"\n", trimmed)
	fmt.Printf("	.zero %d\n", len(init.Content)-len(trimmed))
}

func (init *GVarInitInt) gen(_ types.Type) {
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

func (init *GVarInitZero) gen(_ types.Type) {
	fmt.Printf("	.zero %d\n", init.len)
}

func GenDataGVar(init GVarInit, t types.Type) {
	if init == nil {
		fmt.Printf("	.zero %d\n", t.Size())
	} else {
		init.gen(t)
	}
}
