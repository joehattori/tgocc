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
	name    string
	ty      ty
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

func (t *tokenized) searchStructTag(tag string) *structTag {
	scope := t.curScope
	for scope != nil {
		if tag := scope.searchStructTag(tag); tag != nil {
			return tag
		}
		scope = scope.super
	}
	return nil
}
