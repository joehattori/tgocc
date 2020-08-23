package main

import "log"

type variable interface {
	getName() string
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

func (v *lVar) getName() string { return v.name }
func (v *gVar) getName() string { return v.name }

func (v *lVar) getType() ty { return v.ty }
func (v *gVar) getType() ty { return v.ty }

func newLVar(name string, ty ty) *lVar {
	return &lVar{name: name, ty: ty}
}

func newGVar(name string, ty ty, content interface{}) *gVar {
	return &gVar{name, ty, content}
}

func (t *tokenized) addLVarToScope(id string, ty ty) *lVar {
	if _, islVar := t.curScope.searchVar(id).(*lVar); islVar {
		log.Fatalf("variable %s is already defined", id)
	}
	v := newLVar(id, ty)
	t.curScope.vars = append(t.curScope.vars, v)
	return v
}

func (t *tokenized) findVar(s string) variable {
	v := t.searchVar(s)
	if v == nil {
		log.Fatalf("undefined variable %s", s)
	}
	return v
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

func (s *scope) searchVar(varName string) variable {
	for _, v := range s.vars {
		if v.getName() == varName {
			return v
		}
	}
	return nil
}

func (s *scope) segregateScopeVars() (lVars []*lVar, gVars []*gVar) {
	for _, sv := range s.vars {
		switch v := sv.(type) {
		case *lVar:
			lVars = append(lVars, v)
		case *gVar:
			gVars = append(gVars, v)
		}
	}
	return
}
