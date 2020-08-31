package main

import "log"

type scope struct {
	baseOffset int
	curOffset  int
	super      *scope
	structTags []*structTag
	enumTags   []*enumTag
	vars       []variable
}

func newScope(super *scope) *scope {
	return &scope{baseOffset: super.curOffset, super: super}
}

type structTag struct {
	name string
	ty   ty
}

func newStructTag(name string, t ty) *structTag {
	return &structTag{name, t}
}

type enumTag struct {
	name string
	ty   ty // TODO: is it really necessary?
}

func newEnumTag(name string, t ty) *enumTag {
	return &enumTag{name, t}
}

func (s *scope) addGVar(emit bool, id string, t ty, init gVarInit) *gVar {
	if _, exists := s.searchVar(id).(*gVar); exists {
		log.Fatalf("variable %s is already defined", id)
	}
	v := newGVar(emit, id, t, init)
	s.vars = append(s.vars, v)
	return v
}

func (s *scope) addLVar(id string, t ty) *lVar {
	if _, exists := s.searchVar(id).(*lVar); exists {
		log.Fatalf("variable %s is already defined", id)
	}
	v := newLVar(id, t)
	s.vars = append(s.vars, v)
	return v
}

func (s *scope) addTypeDef(id string, t ty) *typeDef {
	if _, exists := s.searchVar(id).(*typeDef); exists {
		log.Fatalf("typedef %s is already defined", id)
	}
	v := newTypeDef(id, t)
	s.vars = append(s.vars, v)
	return v
}

func (s *scope) addEnum(id string, t ty, val int) *enum {
	if _, exists := s.searchVar(id).(*enum); exists {
		log.Fatalf("enum %s is already defined", id)
	}
	v := newEnum(id, t, val)
	s.vars = append(s.vars, v)
	return v
}

func (s *scope) addStructTag(tag *structTag) {
	if s.searchStructTag(tag.name) != nil {
		log.Fatalf("struct tag %s already exists", tag.name)
	}
	s.structTags = append(s.structTags, tag)
}

func (s *scope) addEnumTag(tag *enumTag) {
	if s.searchEnumTag(tag.name) != nil {
		log.Fatalf("enum tag %s already exists", tag.name)
	}
	s.enumTags = append(s.enumTags, tag)
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

func (s *scope) searchEnumTag(tagStr string) *enumTag {
	for _, tag := range s.enumTags {
		if tag.name == tagStr {
			return tag
		}
	}
	return nil
}

// currently unused
func (s *scope) segregateScopeVars() (lVars []*lVar, gVars []*gVar, typeDefs []*typeDef) {
	for _, sv := range s.vars {
		switch v := sv.(type) {
		case *lVar:
			lVars = append(lVars, v)
		case *gVar:
			gVars = append(gVars, v)
		case *typeDef:
			typeDefs = append(typeDefs, v)
		}
	}
	return
}
