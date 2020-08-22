package main

import "log"

type variable interface {
	getType() ty
}

type lVar struct {
	name   string
	ty     ty
	offset int
}

type gVar struct {
	name string
	ty   ty
	// TODO: elaborate
	content interface{}
}

func (v *lVar) getType() ty { return v.ty }
func (v *gVar) getType() ty { return v.ty }

func newLVar(name string, ty ty, offset int) *lVar {
	return &lVar{name, ty, offset}
}

func newGVar(name string, ty ty, content interface{}) *gVar {
	return &gVar{name, ty, content}
}

func (t *tokenized) buildLVarNode(s string, ty ty, isArg bool) node {
	if _, islVar := t.searchVar(s).(*lVar); islVar {
		log.Fatalf("variable %s is already defined", s)
	}
	f := t.curFn
	offset := f.stackSize + ty.size()
	arg := &lVar{name: s, ty: ty, offset: offset}
	f.lVars = append(f.lVars, arg)
	if isArg {
		f.params = append(f.params, arg)
	}
	// TODO: align
	f.stackSize = offset
	return &nullNode{}
}

func (t *tokenized) findVar(s string) variable {
	v := t.searchVar(s)
	if v == nil {
		log.Fatalf("undefined variable %s", s)
	}
	return v
}

func (t *tokenized) searchVar(varName string) variable {
	for _, lv := range t.curFn.lVars {
		if lv.name == varName {
			return lv
		}
	}
	for _, g := range t.res.gvars {
		if g.name == varName {
			return g
		}
	}
	return nil
}
