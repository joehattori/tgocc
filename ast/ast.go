package ast

import (
	"fmt"

	"github.com/joehattori/tgocc/types"
	"github.com/joehattori/tgocc/vars"
)

type Ast struct {
	Fns   []*FnNode
	GVars []*vars.GVar
}

func (a *Ast) Gen() {
	fmt.Println(".intel_syntax noprefix")
	a.genData()
	a.genText()
}

func (a *Ast) genData() {
	fmt.Println(".data")
	for _, g := range a.GVars {
		if g.Emit {
			fmt.Printf("%s:\n", g.Name())
		}
		genDataGVar(g.Init, g.Type())
	}
}

func (a *Ast) genText() {
	fmt.Println(".text")
	for _, f := range a.Fns {
		f.gen()
	}
}

func genDataGVar(init vars.GVarInit, t types.Type) {
	if init == nil {
		fmt.Printf("	.zero %d\n", t.Size())
	} else {
		init.Gen(t)
	}
}
