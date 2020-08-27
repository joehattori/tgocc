package main

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

type parser struct {
	curFnName string
	curScope  *scope
	res       *ast
	toks      []token
}

func newParser(toks []token) *parser {
	return &parser{"", &scope{}, &ast{}, toks}
}

func (p *parser) spawnScope() {
	p.curScope = newScope(p.curScope)
}

func (p *parser) rewindScope() {
	offset := p.curScope.curOffset
	base := p.curScope.baseOffset
	for _, v := range p.curScope.vars {
		offset = alignTo(offset, v.getType().size())
		offset += v.getType().size()
		if lv, ok := v.(*lVar); ok {
			lv.offset = offset + base
		}
	}
	p.curScope.curOffset += offset
	p.curScope = p.curScope.super
	// maybe this is not necessary if we zero out the memory for child scope?
	p.curScope.curOffset += offset
}

func (p *parser) findVar(s string) variable {
	v := p.searchVar(s)
	if v == nil {
		log.Fatalf("undefined variable %s", s)
	}
	return v
}

func (p *parser) searchStructTag(tag string) *structTag {
	scope := p.curScope
	for scope != nil {
		if tag := scope.searchStructTag(tag); tag != nil {
			return tag
		}
		scope = scope.super
	}
	return nil
}

func (p *parser) searchEnumTag(tag string) *enumTag {
	scope := p.curScope
	for scope != nil {
		if tag := scope.searchEnumTag(tag); tag != nil {
			return tag
		}
		scope = scope.super
	}
	return nil
}

func (p *parser) searchVar(varName string) variable {
	scope := p.curScope
	for scope != nil {
		if v := scope.searchVar(varName); v != nil {
			return v
		}
		scope = scope.super
	}
	return nil
}

/*
   Prerequisite functions for parsing.
*/

func (p *parser) beginsWith(s string) bool {
	return strings.HasPrefix(p.toks[0].getStr(), s)
}

func (p *parser) consume(str string) bool {
	if r, ok := p.toks[0].(*reservedTok); ok &&
		r.len == len(str) &&
		strings.HasPrefix(r.str, str) {
		p.popToks()
		return true
	}
	return false
}

func (p *parser) consumeID() (string, bool) {
	cur := p.toks[0]
	id := idRegexp.FindString(cur.getStr())
	if _, ok := cur.(*idTok); ok && id != "" {
		p.popToks()
		return id, true
	}
	return "", false
}

func (p *parser) consumeStr() (string, bool) {
	if s, ok := p.toks[0].(*strTok); ok {
		p.popToks()
		return s.content, true
	}
	return "", false
}

func (p *parser) expect(str string) {
	cur := p.toks[0]
	if r, ok := cur.(*reservedTok); ok &&
		r.len == len(str) && strings.HasPrefix(cur.getStr(), str) {
		p.popToks()
		return
	}
	log.Fatalf("%s was expected but got %s", str, cur.getStr())
}

func (p *parser) expectID() string {
	cur := p.toks[0]
	id := idRegexp.FindString(cur.getStr())
	if _, ok := cur.(*idTok); ok && id != "" {
		p.popToks()
		return id
	}
	log.Fatalf("ID was expected but got %s", cur.getStr())
	return ""
}

func (p *parser) expectNum() int {
	cur := p.toks[0]
	if n, ok := cur.(*numTok); ok {
		p.popToks()
		return n.val
	}
	log.Fatalf("Number was expected but got %s", cur.getStr())
	return -1
}

func (p *parser) isFunction() bool {
	orig := p.toks
	defer func() { p.toks = orig }()
	p.baseType()
	for p.consume("*") {
	}
	_, isID := p.consumeID()
	return isID && p.consume("(")
}

func (p *parser) isType() bool {
	cur := p.toks[0]
	switch tok := cur.(type) {
	case *idTok:
		id := idRegexp.FindString(tok.id)
		_, ok := p.searchVar(id).(*typeDef)
		return ok
	case *reservedTok:
		return typeRegexp.MatchString(tok.str)
	}
	return false
}

func (p *parser) popToks() {
	p.toks = p.toks[1:]
}

var gVarLabelCount int

func newGVarLabel() string {
	defer func() { gVarLabelCount++ }()
	return fmt.Sprintf(".L.data.%d", gVarLabelCount)
}

/*
   Actual parsing process from here.

   program    = (function | globalVar)*
   function   = baseType tyDecl "(" (ident ("," ident)* )? ")" ("{" stmt* "}" | ";")
   globalVar  = decl
   stmt       = expr ";"
  				| "{" stmt* "}"
  				| "return" expr ";"
  				| "if" "(" expr ")" stmt ("else" stmt) ?
  				| "while" "(" expr ")" stmt
  				| "for" "(" expr? ";" expr? ";" expr? ")" stmt
  				| "typedef" ty ident ("[" num "]")* ";"
  				| decl
   decl       = baseType tyDecl ("[" expr "]")* "=" expr ;" | baseTy ";"
   tyDecl     = "*"* (ident | "(" tyDecl ")")
   expr       = assign
   stmtExpr   = "(" "{" stmt+ "}" ")"
   assign     = equality ("=" assign) ?
   equality   = relational ("==" relational | "!=" relational)*
   relational = add ("<" add | "<=" add | ">" add | ">=" add)*
   add        = mul ("+" mul | "-" mul)*
   mul        = cat ("*" cast | "/" cast)*
   cast       = "(" baseType "*"*  ")" cast | unary
   unary      = ("+" | "-" | "*" | "&")? cast | postfix
   postfix    = primary (("[" expr "]") | ("." | id))*
   primary    =  num
  				| "sizeof" unary
  				| str
  				| ident ("(" (expr ("," expr)* )? ")")?
  				| "(" expr ")"
  				| stmtExpr
*/

func (p *parser) parse() {
	ast := p.res
	for {
		if _, ok := p.toks[0].(*eofTok); ok {
			break
		}
		if p.isFunction() {
			fn := p.function()
			if fn != nil {
				ast.fns = append(ast.fns, fn)
			}
		} else {
			// TODO: gvar init
			ty, id, _ := p.decl()
			if ty != nil {
				p.curScope.addGVar(id, ty, nil)
				ast.gVars = append(ast.gVars, newGVar(id, ty, nil))
			}
		}
	}
}

func (p *parser) function() *fnNode {
	ty, _ := p.baseType()
	fnName, ty := p.tyDecl(ty)
	p.curFnName = fnName
	p.curScope.addGVar(fnName, newTyFn(ty), nil)
	fn := newFnNode(fnName, ty)
	p.spawnScope()
	p.readFnParams(fn)
	if p.consume(";") {
		p.rewindScope()
		return nil
	}
	p.expect("{")
	for !p.consume("}") {
		fn.body = append(fn.body, p.stmt())
	}
	p.setFnLVars(fn)
	p.rewindScope()
	// TODO: align
	fn.stackSize = p.curScope.curOffset
	return fn
}

func (p *parser) decl() (t ty, id string, rhs node) {
	t, isTypeDef := p.baseType()
	if p.consume(";") {
		return
	}
	id, t = p.tyDecl(t)
	if isTypeDef {
		p.expect(";")
		p.curScope.addTypeDef(id, t)
		// returned t is nil when it is typedef (no need to add to scope.vars)
		return nil, "", nil
	}
	t = p.tySuffix(t)
	if p.consume(";") {
		return
	}
	p.expect("=")
	rhs = p.expr()
	p.expect(";")
	return
}

func (p *parser) baseType() (t ty, isTypeDef bool) {
	if p.consume("typedef") {
		isTypeDef = true
	}
	cur := p.toks[0]
	switch tok := cur.(type) {
	case *idTok:
		id := idRegexp.FindString(tok.id)
		if tyDef, ok := p.searchVar(id).(*typeDef); ok {
			p.popToks()
			return tyDef.ty, isTypeDef
		}
	case *reservedTok:
		if p.beginsWith("struct") {
			return p.structDecl(), isTypeDef
		}
		if p.beginsWith("enum") {
			return p.enumDecl(), isTypeDef
		}
		if p.consume("int") {
			return newTyInt(), isTypeDef
		}
		if p.consume("char") {
			return newTyChar(), isTypeDef
		}
		if p.consume("short") {
			p.consume("int")
			return newTyShort(), isTypeDef
		}
		if p.consume("long") {
			p.consume("long")
			p.consume("int")
			return newTyLong(), isTypeDef
		}
		if p.consume("void") {
			return newTyVoid(), isTypeDef
		}
		if p.consume("_Bool") {
			return newTyBool(), isTypeDef
		}
	}
	log.Fatalf("type expected but got %T: %s", cur, cur.getStr())
	return
}

func (p *parser) tyDecl(baseTy ty) (id string, newTy ty) {
	for p.consume("*") {
		baseTy = newTyPtr(baseTy)
	}
	if p.consume("(") {
		id, newTy = p.tyDecl(nil)
		p.expect(")")
		baseTy = p.tySuffix(baseTy)
		switch t := newTy.(type) {
		case *tyArr:
			t.of = baseTy
		case *tyPtr:
			t.to = baseTy
		default:
			newTy = baseTy
		}
		return
	}
	return p.expectID(), p.tySuffix(baseTy)
}

func (p *parser) tySuffix(baseTy ty) ty {
	if p.consume("[") {
		l := p.expectNum()
		p.expect("]")
		baseTy = p.tySuffix(baseTy)
		return newTyArr(baseTy, l)
	}
	return baseTy
}

func (p *parser) readFnParams(fn *fnNode) {
	p.expect("(")
	isFirstArg := true
	for !p.consume(")") {
		if !isFirstArg {
			p.expect(",")
		}
		isFirstArg = false

		ty, _ := p.baseType()
		id, ty := p.tyDecl(ty)
		lv := p.curScope.addLVar(id, ty)
		fn.params = append(fn.params, lv)
	}
}

func (p *parser) setFnLVars(fn *fnNode) {
	offset := 0
	for _, sv := range p.curScope.vars {
		switch v := sv.(type) {
		case *lVar:
			offset = alignTo(offset, v.getType().alignment())
			offset += v.getType().size()
			v.offset = offset
			fn.lVars = append(fn.lVars, v)
		}
	}
}

func (p *parser) structDecl() ty {
	p.expect("struct")
	tagStr, tagExists := p.consumeID()
	if tagExists && !p.beginsWith("{") {
		if tag := p.searchStructTag(tagStr); tag != nil {
			return tag.ty
		}
		log.Fatalf("no such struct tag %s", tagStr)
	}
	p.expect("{")
	var members []*member
	offset, align := 0, 0
	for !p.consume("}") {
		// TODO: handle when rhs is not null
		ty, tag, _ := p.decl()
		offset = alignTo(offset, ty.alignment())
		members = append(members, newMember(tag, offset, ty))
		offset += ty.size()
		if align < ty.size() {
			align = ty.size()
		}
	}
	tyStruct := newTyStruct(align, members, alignTo(offset, align))
	if tagExists {
		p.curScope.addStructTag(newStructTag(tagStr, tyStruct))
	}
	return tyStruct
}

func (p *parser) enumDecl() ty {
	p.expect("enum")
	tagStr, tagExists := p.consumeID()
	if tagExists && !p.beginsWith("{") {
		if tag := p.searchEnumTag(tagStr); tag != nil {
			return tag.ty
		}
		log.Fatalf("no such enum tag %s", tagStr)
	}
	t := newTyEnum()

	p.expect("{")
	c := 0
	for {
		id := p.expectID()
		if p.consume("=") {
			c = p.expectNum()
		}
		p.curScope.addEnum(id, t, c)
		c++
		orig := p.toks
		if p.consume("}") || p.consume(",") && p.consume("}") {
			break
		}
		p.toks = orig
		p.expect(",")
	}
	if tagExists {
		p.curScope.addEnumTag(newEnumTag(tagStr, t))
	}
	return t
}

func (p *parser) stmt() node {
	// handle block
	if p.consume("{") {
		var blkStmts []node
		p.spawnScope()
		for !p.consume("}") {
			blkStmts = append(blkStmts, p.stmt())
		}
		p.rewindScope()
		return newBlkNode(blkStmts)
	}

	// handle return
	if p.consume("return") {
		node := newRetNode(p.expr(), p.curFnName)
		p.expect(";")
		return node
	}

	// handle if statement
	if p.consume("if") {
		p.expect("(")
		cond := p.expr()
		p.expect(")")
		then := p.stmt()

		var els node
		if p.consume("else") {
			els = p.stmt()
		}

		return newIfNode(cond, then, els)
	}

	// handle while statement
	if p.consume("while") {
		p.expect("(")
		condNode := p.expr()
		p.expect(")")
		thenNode := p.stmt()
		return newWhileNode(condNode, thenNode)
	}

	// handle for statement
	if p.consume("for") {
		p.expect("(")

		var init, cond, inc, then node

		if !p.consume(";") {
			init = newExprNode(p.expr())
			p.expect(";")
		}

		if !p.consume(";") {
			cond = p.expr()
			p.expect(";")
		}

		if !p.consume(")") {
			inc = newExprNode(p.expr())
			p.expect(")")
		}

		then = p.stmt()
		return newForNode(init, cond, inc, then)
	}

	// handle variable definition
	if p.isType() {
		ty, id, rhs := p.decl()
		if id == "" {
			return newNullNode()
		}
		p.curScope.addLVar(id, ty)
		if rhs == nil {
			return newNullNode()
		}
		node := newAssignNode(newVarNode(p.findVar(id)), rhs)
		return newExprNode(node)
	}

	node := p.expr()
	p.expect(";")
	return newExprNode(node)
}

func (p *parser) expr() node {
	return p.assign()
}

func (p *parser) assign() node {
	node := p.equality()
	if p.consume("=") {
		assignNode := p.assign()
		node = newAssignNode(node.(addressableNode), assignNode)
	}
	return node
}

func (p *parser) equality() node {
	node := p.relational()
	for {
		if p.consume("==") {
			node = newArithNode(ndEq, node, p.relational())
		} else if p.consume("!=") {
			node = newArithNode(ndNeq, node, p.relational())
		} else {
			return node
		}
	}
}

func (p *parser) relational() node {
	node := p.addSub()
	for {
		if p.consume("<=") {
			node = newArithNode(ndLeq, node, p.addSub())
		} else if p.consume(">=") {
			node = newArithNode(ndGeq, node, p.addSub())
		} else if p.consume("<") {
			node = newArithNode(ndLt, node, p.addSub())
		} else if p.consume(">") {
			node = newArithNode(ndGt, node, p.addSub())
		} else {
			return node
		}
	}
}

func (p *parser) addSub() node {
	node := p.mulDiv()
	for {
		if p.consume("+") {
			node = newAddNode(node, p.mulDiv())
		} else if p.consume("-") {
			node = newSubNode(node, p.mulDiv())
		} else {
			return node
		}
	}
}

func (p *parser) mulDiv() node {
	node := p.cast()
	for {
		if p.consume("*") {
			node = newArithNode(ndMul, node, p.cast())
		} else if p.consume("/") {
			node = newArithNode(ndDiv, node, p.cast())
		} else {
			return node
		}
	}
}

func (p *parser) cast() node {
	orig := p.toks
	if p.consume("(") {
		if p.isType() {
			t, _ := p.baseType()
			for p.consume("*") {
				t = newTyPtr(t)
			}
			p.expect(")")
			return newCastNode(p.cast(), t)
		}
		p.toks = orig
	}
	return p.unary()
}

func (p *parser) unary() node {
	if p.consume("+") {
		return p.cast()
	}
	if p.consume("-") {
		node := p.cast()
		return newSubNode(newNumNode(0), node)
	}
	if p.consume("*") {
		node := p.cast()
		return newDerefNode(node)
	}
	if p.consume("&") {
		node := p.cast()
		return newAddrNode(node.(addressableNode))
	}
	return p.postfix()
}

func (p *parser) postfix() node {
	node := p.primary()
	for {
		if p.consume("[") {
			add := newAddNode(node, p.expr())
			node = newDerefNode(add)
			p.expect("]")
			continue
		}
		if p.consume(".") {
			if s, ok := node.loadType().(*tyStruct); ok {
				mem := s.findMember(p.expectID())
				node = newMemberNode(node.(addressableNode), mem)
				continue
			}
			log.Fatalf("expected struct but got %T", node.loadType())
		}
		if p.consume("->") {
			if t, ok := node.loadType().(*tyPtr); ok {
				mem := t.to.(*tyStruct).findMember(p.expectID())
				node = newMemberNode(newDerefNode(node.(addressableNode)), mem)
				continue
			}
			log.Fatalf("expected struct but got %T", node.loadType())
		}
		return node
	}
}

func (p *parser) stmtExpr() node {
	// "(" and "{" is already read.
	p.spawnScope()
	body := make([]node, 0)
	body = append(body, p.stmt())
	for !p.consume("}") {
		body = append(body, p.stmt())
	}
	p.expect(")")
	if ex, ok := body[len(body)-1].(*exprNode); !ok {
		log.Fatal("statement expression returning void is not supported")
	} else {
		body[len(body)-1] = ex.body
	}
	p.rewindScope()
	return newStmtExprNode(body)
}

func (p *parser) primary() node {
	if p.consume("(") {
		if p.consume("{") {
			return p.stmtExpr()
		}
		node := p.expr()
		p.expect(")")
		return node
	}

	if p.consume("sizeof") {
		orig := p.toks
		if p.consume("(") {
			if p.isType() {
				base, _ := p.baseType()
				n := base.size()
				p.expect(")")
				return newNumNode(int64(n))
			}
			p.toks = orig
		}
		return newNumNode(int64(p.unary().loadType().size()))
	}

	if id, isID := p.consumeID(); isID {
		if p.consume("(") {
			var t ty
			if fn, ok := p.searchVar(id).(*gVar); ok {
				t = fn.ty.(*tyFn).retTy
			} else {
				t = newTyInt()
			}
			var params []node
			if p.consume(")") {
				return newFnCallNode(id, params, t)
			}
			params = append(params, p.expr())
			for p.consume(",") {
				params = append(params, p.expr())
			}
			p.expect(")")
			return newFnCallNode(id, params, t)
		}

		switch v := p.findVar(id).(type) {
		case *enum:
			return newNumNode(int64(v.val))
		case *lVar, *gVar:
			return newVarNode(v)
		default:
			log.Fatalf("unhandled case of variable in primary: %T", p.findVar(id))
		}
	}

	if str, isStr := p.consumeStr(); isStr {
		s := newGVar(newGVarLabel(), newTyArr(newTyChar(), utf8.RuneCountInString(str)), str)
		p.res.gVars = append(p.res.gVars, s)
		return newVarNode(s)
	}

	return newNumNode(int64(p.expectNum()))
}
