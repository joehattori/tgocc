package parser

import (
	"log"

	"github.com/joehattori/tgocc/types"
	"github.com/joehattori/tgocc/vars"
)

type scope struct {
	baseOffset int
	curOffset  int
	super      *scope
	structTags []*structTag
	enumTags   []*enumTag
	vars       []vars.Var
}

func newScope(super *scope) *scope {
	return &scope{baseOffset: super.curOffset, super: super}
}

type structTag struct {
	name string
	ty   types.Type
}

func newStructTag(name string, t types.Type) *structTag {
	return &structTag{name, t}
}

type enumTag struct {
	name string
	ty   types.Type // TODO: is it really necessary?
}

func newEnumTag(name string, t types.Type) *enumTag {
	return &enumTag{name, t}
}

func (s *scope) addGVar(emit bool, id string, t types.Type, init vars.GVarInit) *vars.GVar {
	if _, exists := s.searchVar(id).(*vars.GVar); exists {
		log.Fatalf("identifier %s is already defined", id)
	}
	v := vars.NewGVar(emit, id, t, init)
	s.vars = append(s.vars, v)
	return v
}

func (s *scope) addLVar(id string, t types.Type) *vars.LVar {
	if _, exists := s.searchVar(id).(*vars.LVar); exists {
		log.Fatalf("identifier %s is already defined", id)
	}
	v := vars.NewLVar(id, t)
	s.vars = append(s.vars, v)
	return v
}

func (s *scope) addTypeDef(id string, t types.Type) *vars.TypeDef {
	if _, exists := s.searchVar(id).(*vars.TypeDef); exists {
		log.Fatalf("typedef %s is already defined", id)
	}
	v := vars.NewTypeDef(id, t)
	s.vars = append(s.vars, v)
	return v
}

func (s *scope) addEnum(id string, t types.Type, val int) *vars.Enum {
	if _, exists := s.searchVar(id).(*vars.Enum); exists {
		log.Fatalf("enum %s is already defined", id)
	}
	v := vars.NewEnum(id, t, val)
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

func (s *scope) searchVar(varName string) vars.Var {
	for _, v := range s.vars {
		if v.Name() == varName {
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

func (s *scope) segregateScopeVars() (lVars []*vars.LVar, gVars []*vars.GVar, typeDefs []*vars.TypeDef) {
	for _, sv := range s.vars {
		switch v := sv.(type) {
		case *vars.LVar:
			lVars = append(lVars, v)
		case *vars.GVar:
			gVars = append(gVars, v)
		case *vars.TypeDef:
			typeDefs = append(typeDefs, v)
		}
	}
	return
}
