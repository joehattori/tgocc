package main

import "log"

type scope struct {
	baseOffset int
	curOffset  int
	super      *scope
	vars       []variable
	structTags []*structTag
}

type structTag struct {
	name string
	ty   ty
}

func newScope(super *scope) *scope {
	return &scope{super.curOffset, 0, super, nil, nil}
}

func newStructTag(name string, ty ty) *structTag {
	return &structTag{name, ty}
}

func (s *scope) addLVar(id string, ty ty) *lVar {
	if _, exists := s.searchVar(id).(*lVar); exists {
		log.Fatalf("variable %s is already defined", id)
	}
	v := newLVar(id, ty)
	s.vars = append(s.vars, v)
	return v
}

func (s *scope) addStructTag(tag *structTag) {
	if s.searchStructTag(tag.name) != nil {
		log.Fatalf("struct tag %s already exists", tag.name)
	}
	s.structTags = append(s.structTags, tag)
}

func (s *scope) searchVar(varName string) variable {
	for _, v := range s.vars {
		if v.getName() == varName {
			return v
		}
	}
	return nil
}

func (s *scope) searchStructTag(tagStr string) *structTag {
	for _, tag := range s.structTags {
		if tag.name == tagStr {
			return tag
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
