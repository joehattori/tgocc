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
		r.len == utf8.RuneCountInString(str) &&
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
		r.len == utf8.RuneCountInString(str) && strings.HasPrefix(cur.getStr(), str) {
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

func (p *parser) expectNum() int64 {
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
  				| "for" "(" (expr? ";" | decl) expr? ";" expr? ")" stmt
				| "typedef" ty ident ("[" constExpr "]")* ";"
				| "switch" "(" expr ")" "{" switchCase* ("default" ":" stmt*)? }"
  				| decl
	switchCase = "case" num ":" stmt*
	decl       = baseType tyDecl ("[" constExpr "]")* "=" initialize ;" | baseTy ";"
	tyDecl     = "*"* (ident | "(" tyDecl ")")
	expr       = assign
	assign     = ternary (("=" | "+=" | "-=" | "*=" | "/=") assign) ?
	ternary    = logOr ("?" expr ":" ternary)?
	logOr      = logAnd ("||" logAnd)*
	logAnd     = bitOr ("&&" bitOr)*
	bitOr      = bitXor ("|" bitXor)*
	bitXor     = bitAnd ("^" bitAnd)*
	bitAnd     = equality ("&" equality)*
	equality   = relational ("==" relational | "!=" relational)*
	relational = shift ("<" shift | "<=" shift | ">" shift | ">=" shift)*
	shift      = add ("<<" add | ">>" add | "<<=" add | ">>=" add)
	add        = mul ("+" mul | "-" mul)*
	mul        = cat ("*" cast | "/" cast)*
	cast       = "(" baseType "*"*  ")" cast | unary
	unary      = ("+" | "-" | "*" | "&" | "!" | "~")? cast | ("++" | "--") unary | postfix
	postfix    = primary (("[" expr "]") | ("." ident) | ("->" ident) | "++" | "--")*
	primary    =  num
				| "sizeof" unary
				| str
				| ident ("(" (expr ("," expr)* )? ")")?
				| "(" expr ")"
				| stmtExpr
	stmtExpr   = "(" "{" stmt+ "}" ")"
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
			// global variable initialization
			ty, id, rhs := p.decl()
			if ty != nil {
				init := buildGVarInit(ty, rhs)
				p.curScope.addGVar(id, ty, init)
				ast.gVars = append(ast.gVars, newGVar(id, ty, init))
			}
		}
	}
}

func buildGVarInit(ty ty, rhs node) gVarInit {
	if rhs == nil {
		return nil
	}
	switch t := ty.(type) {
	case *tyArr:
		var body []gVarInit
		idx := 0
		for _, e := range rhs.(*blkNode).body {
			idx++
			body = append(body, buildGVarInit(t.of, e))
		}
		if t.len < 0 {
			t.len = idx
		}
		// zero out the rest
		if _, ok := t.of.(*tyArr); !ok {
			for i := idx; i < t.len; i++ {
				body = append(body, newGVarInitInt(0, t.of.size()))
			}
		}
		return newGVarInitArr(body)
	case *tyStruct:
		var body []gVarInit
		idx := 0
		for i, e := range rhs.(*blkNode).body {
			idx++
			mem := t.members[i]
			var toAppend []gVarInit
			toAppend = append(toAppend, buildGVarInit(mem.ty, e))
			var end int
			if i < len(t.members) - 1 {
				end = t.members[i+1].offset
			} else {
				end = t.size()
			}
			for j := mem.offset + mem.ty.size(); j < end; j++ {
				toAppend = append(toAppend, newGVarInitInt(0, 1))
			}
			body = append(body, newGVarInitArr(toAppend))
		}
		// zero out the rest
		for i := idx; i < len(t.members); i++ {
			body = append(body, newGVarInitInt(0, t.members[i].ty.size()))
		}
		return newGVarInitArr(body)
	default:
		switch nd := rhs.(type) {
		case *addrNode:
			return newGVarInitLabel(nd.v.(*varNode).v.getName())
		case *varNode:
			if _, ok := nd.v.getType().(*tyArr); ok {
				return newGVarInitLabel(nd.v.getName())
			}
		}
		return newGVarInitInt(eval(rhs), t.size())
	}
}

func (p *parser) function() *fnNode {
	ty, _, sc := p.baseType()
	fnName, ty := p.tyDecl(ty)
	p.curFnName = fnName
	p.curScope.addGVar(fnName, newTyFn(ty), nil)
	fn := newFnNode((sc & static)!= 0, fnName, ty)
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
	t, isTypeDef, _ := p.baseType()
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
	rhs = p.initializer(t)
	p.expect(";")
	return
}

func (p *parser) initializer(t ty) node {
	switch t := t.(type) {
	case *tyArr, *tyStruct:
		var nodes []node
		if str, ok := p.consumeStr(); ok {
			init := newGVarInitStr(str)
			s := newGVar(newGVarLabel(), newTyArr(newTyChar(), len(str)), init)
			p.res.gVars = append(p.res.gVars, s)
			return newVarNode(s)
		}
		p.expect("{")
		for !p.consume("}") {
			switch t := t.(type) {
			case *tyArr:
				nodes = append(nodes, p.initializer(t.of))
			case *tyStruct:
				nodes = append(nodes, p.assign())
			}
			if !p.consume(",") {
				p.expect("}")
				break
			}
		}
		return newBlkNode(nodes)
	}
	return p.assign()
}

type storageClass int

const (
	static = 0b01
	extern = 0b10 // TODO
)

func (p *parser) baseType() (t ty, isTypeDef bool, sc storageClass) {
	if p.consume("typedef") {
		isTypeDef = true
	}
	if p.consume("static")  {
		sc |= static
	}
	cur := p.toks[0]
	switch tok := cur.(type) {
	case *idTok:
		id := idRegexp.FindString(tok.id)
		if tyDef, ok := p.searchVar(id).(*typeDef); ok {
			p.popToks()
			return tyDef.ty, isTypeDef, sc
		}
	case *reservedTok:
		if p.beginsWith("struct") {
			return p.structDecl(), isTypeDef, sc
		}
		if p.beginsWith("enum") {
			return p.enumDecl(), isTypeDef, sc
		}
		if p.consume("int") {
			return newTyInt(), isTypeDef,sc
		}
		if p.consume("char") {
			return newTyChar(), isTypeDef,sc
		}
		if p.consume("short") {
			p.consume("int")
			return newTyShort(), isTypeDef,sc
		}
		if p.consume("long") {
			p.consume("long")
			p.consume("int")
			return newTyLong(), isTypeDef,sc
		}
		if p.consume("void") {
			return newTyVoid(), isTypeDef,sc
		}
		if p.consume("_Bool") {
			return newTyBool(), isTypeDef,sc
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

func (p *parser) tySuffix(t ty) ty {
	if !p.consume("[") {
		return t
	}
	l := -1
	if !p.consume("]") {
		l = int(p.constExpr())
		p.expect("]")
	}
	t = p.tySuffix(t)
	return newTyArr(t, l)
}

func (p *parser) constExpr() int64 {
	return eval(p.ternary())
}

func eval(nd node) int64 {
	switch n := nd.(type){
	case *arithNode:
		switch n.op {
		case ndAdd:
			return eval(n.lhs) + eval(n.rhs)
		case ndSub:
			return eval(n.lhs) - eval(n.rhs)
		case ndMul:
			return eval(n.lhs) * eval(n.rhs)
		case ndDiv:
			return eval(n.lhs) / eval(n.rhs)
		case ndBitOr:
			return eval(n.lhs) | eval(n.rhs)
		case ndBitXor:
			return eval(n.lhs) ^ eval(n.rhs)
		case ndBitAnd:
			return eval(n.lhs) & eval(n.rhs)
		case ndShl:
			return eval(n.lhs) << eval(n.rhs)
		case ndShr:
			return eval(n.lhs) >> eval(n.rhs)
		case ndEq:
			if eval(n.lhs) == eval(n.rhs) {
				return 1
			}
			return 0
		case ndNeq:
			if eval(n.lhs) != eval(n.rhs) {
				return 1
			}
			return 0
		case ndLt:
			if eval(n.lhs) < eval(n.rhs) {
				return 1
			}
			return 0
		case ndLeq:
			if eval(n.lhs) <= eval(n.rhs) {
				return 1
			}
			return 0
		case ndGt:
			if eval(n.lhs) > eval(n.rhs) {
				return 1
			}
			return 0
		case ndGeq:
			if eval(n.lhs) <= eval(n.rhs) {
				return 1
			}
			return 0
		case ndLogAnd:
			return eval(n.lhs) & eval(n.rhs)
		case ndLogOr:
			return eval(n.lhs) | eval(n.rhs)
		}
	case *bitNotNode:
		return ^eval(n.body)
	case *notNode:
		if eval(n.body) != 0 {
			return 1
		}
		return 0
	case *numNode:
		return n.val
	case *ternaryNode:
		if eval(n.cond) == 0 {
			return eval(n.rhs)
		}
		return eval(n.lhs)
	}
	log.Fatalf("not a constant expression.")
	return 0
}

func (p *parser) readFnParams(fn *fnNode) {
	p.expect("(")
	isFirstArg := true
	for !p.consume(")") {
		if !isFirstArg {
			p.expect(",")
		}
		isFirstArg = false

		ty, _, _ := p.baseType()
		id, ty := p.tyDecl(ty)
		if arr, ok := ty.(*tyArr); ok {
			ty = newTyPtr(arr.of)
		}
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
			c = int(p.constExpr())
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
		if p.consume(";") {
			return newRetNode(nil, p.curFnName)
		}
		node := newRetNode(p.expr(), p.curFnName)
		p.expect(";")
		return node
	}

	// handle break
	if p.consume("break") {
		p.expect(";")
		return newBreakNode()
	}

	// handle continue
	if p.consume("continue") {
		p.expect(";")
		return newContinueNode()
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

		p.spawnScope()
		if !p.consume(";") {
			if p.isType() {
				t, id, rhs := p.decl()
				p.curScope.addLVar(id, t)
				if rhs == nil {
					init = newNullNode()
				} else {
					a := newAssignNode(newVarNode(p.findVar(id)), rhs)
					init = newExprNode(a)
				}
			} else {
				init = newExprNode(p.expr())
				p.expect(";")
			}
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
		p.rewindScope()
		return newForNode(init, cond, inc, then)
	}

	// handle switch statement
	if p.consume("switch") {
		p.expect("(")
		e := p.expr()
		p.expect(")")
		p.expect("{")

		var cases []*caseNode
		var dflt *defaultNode = nil
		for idx := 0;; idx++ {
			if c := p.switchCase(idx); c == nil {
				break
			} else {
				switch n := c.(type) {
				case *caseNode:
					cases = append(cases, n)
				case *defaultNode:
					if dflt != nil {
						log.Fatal("multiple definition of default clause.")
					}
					dflt = n
				default:
					log.Fatalf("unhandled case: %T", c)
				}
			}
		}
		p.expect("}")
		return newSwitchNode(e, cases, dflt)
	}

	// handle variable definition
	if p.isType() {
		t, id, rhs := p.decl()
		if id == "" {
			return newNullNode()
		}
		p.curScope.addLVar(id, t)
		if rhs == nil {
			return newNullNode()
		}
		return p.storeInit(t, newVarNode(p.findVar(id)), rhs)
	}

	node := p.expr()
	p.expect(";")
	return newExprNode(node)
}

func (p *parser) storeInit(t ty, dst addressableNode, rhs node) node {
	switch t := t.(type) {
	case *tyArr:
		// TODO: clean up
		var body []node
		var ln, idx int
		_, isChar := t.of.(*tyChar)
		isString, str := isNodeStr(rhs)
		// string literal
		if isChar && isString {
			for i, r := range str {
				idx++
				addr := newDerefNode(newAddNode(dst, newNumNode(int64(i))))
				body = append(body, newExprNode(newAssignNode(addr, newNumNode(int64(r)))))
			}
			ln = len(str)
		} else {
			for i, mem := range rhs.(*blkNode).body {
				idx++
				addr := newDerefNode(newAddNode(dst, newNumNode(int64(i))))
				body = append(body, p.storeInit(t.of, addr, mem))
			}
			ln = len(rhs.(*blkNode).body)
		}

		if t.len < 0 {
			t.len = ln
		}

		if _, ok := t.of.(*tyArr); !ok {
			// zero out on initialization
			for i := idx; i < ln; i++ {
				addr := newDerefNode(newAddNode(dst, newNumNode(int64(i))))
				body = append(body, p.storeInit(t.of, addr, newNumNode(0)))
			}
		}

		return newBlkNode(body)
	case *tyStruct:
		var body []node
		idx := 0
		for i, e := range rhs.(*blkNode).body {
			idx++
			node := newMemberNode(dst, t.members[i])
			body = append(body, newExprNode(newAssignNode(node, e)))
		}
		// zero out the rest
		for i := idx; i < len(t.members); i++ {
			node := newMemberNode(dst, t.members[i])
			body = append(body, newExprNode(newAssignNode(node, newNumNode(0))))
		}
		return newBlkNode(body)
	default:
		return newExprNode(newAssignNode(dst, rhs))
	}
}

func isNodeStr(n node) (bool, string) {
	if v, ok := n.(*varNode); !ok {
		return false, ""
	} else if g, ok := v.v.(*gVar); !ok {
		return false, ""
	} else if t, ok := g.ty.(*tyArr); !ok {
		return false, ""
	} else if _, ok := t.of.(*tyChar); !ok {
		return false, ""
	} else {
		return true, g.init.(*gVarInitStr).content
	}
}

func (p *parser) switchCase(idx int) node {
	if p.consume("case") {
		n := p.constExpr()
		p.expect(":")
		var body []node
		for !p.beginsWith("case") && !p.beginsWith("default") && !p.beginsWith("}") {
			body = append(body, p.stmt())
		}
		return newCaseNode(int(n), body, idx)
	}
	if p.consume("default") {
		p.expect(":")
		var body []node
		for !p.beginsWith("case") && !p.beginsWith("default") && !p.beginsWith("}") {
			body = append(body, p.stmt())
		}
		return newDefaultNode(body, idx)
	}
	return nil
}

func (p *parser) expr() node {
	return p.assign()
}

func (p *parser) assign() node {
	node := p.ternary()
	if p.consume("=") {
		node = newAssignNode(node.(addressableNode), p.assign())
	} else if p.consume("+=") {
		if _, ok := node.loadType().(ptr); ok {
			node = newArithNode(ndPtrAddEq, node.(addressableNode), p.assign())
		} else {
			node = newArithNode(ndAddEq, node.(addressableNode), p.assign())
		}
	} else if p.consume("-=") {
		if _, ok := node.loadType().(ptr); ok {
			node = newArithNode(ndPtrSubEq, node.(addressableNode), p.assign())
		} else {
			node = newArithNode(ndSubEq, node.(addressableNode), p.assign())
		}
	} else if p.consume("*=") {
		node = newArithNode(ndMulEq, node.(addressableNode), p.assign())
	} else if p.consume("/=") {
		node = newArithNode(ndDivEq, node.(addressableNode), p.assign())
	}
	return node
}

func (p *parser) ternary() node {
	node := p.logOr()
	if !p.consume("?") {
		return node
	}
	l := p.expr()
	p.expect(":")
	r := p.ternary()
	return newTernaryNode(node, l, r)
}

func (p *parser) logOr() node {
	node := p.logAnd()
	for p.consume("||") {
		node = newArithNode(ndLogOr, node, p.logAnd())
	}
	return node
}

func (p *parser) logAnd() node {
	node := p.bitOr()
	for p.consume("&&") {
		node = newArithNode(ndLogAnd, node, p.bitOr())
	}
	return node
}

func (p *parser) bitOr() node {
	node := p.bitXor()
	for p.consume("|") {
		node = newArithNode(ndBitOr, node, p.bitXor())
	}
	return node
}

func (p *parser) bitXor() node {
	node := p.bitAnd()
	for p.consume("^") {
		node = newArithNode(ndBitXor, node, p.bitXor())
	}
	return node
}

func (p *parser) bitAnd() node {
	node := p.equality()
	for p.consume("&") {
		node = newArithNode(ndBitAnd,node, p.equality())
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
	node := p.shift()
	for {
		if p.consume("<=") {
			node = newArithNode(ndLeq, node, p.shift())
		} else if p.consume(">=") {
			node = newArithNode(ndGeq, node, p.shift())
		} else if p.consume("<") {
			node = newArithNode(ndLt, node, p.shift())
		} else if p.consume(">") {
			node = newArithNode(ndGt, node, p.shift())
		} else {
			return node
		}
	}
}

func (p *parser) shift() node {
	node := p.addSub()
	for {
		if p.consume("<<") {
			node = newArithNode(ndShl, node, p.shift())
		} else if p.consume(">>") {
			node = newArithNode(ndShr, node, p.shift())
		} else if p.consume("<<=") {
			node = newArithNode(ndShlEq, node, p.shift())
		} else if p.consume(">>=") {
			node = newArithNode(ndShrEq, node, p.shift())
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
			t, _,_ := p.baseType()
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
		return newSubNode(newNumNode(0), p.cast())
	}
	if p.consume("*") {
		return newDerefNode(p.cast())
	}
	if p.consume("&") {
		return newAddrNode(p.cast().(addressableNode))
	}
	if p.consume("!") {
		return newNotNode(p.cast())
	}
	if p.consume("~") {
		return newBitNotNode(p.cast())
	}
	if p.consume("++") {
		return newIncNode(p.unary().(addressableNode), true)
	}
	if p.consume("--") {
		return newDecNode(p.unary().(addressableNode), true)
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
			log.Fatalf("expected pointer but got %T", node.loadType())
		}
		if p.consume("++") {
			node = newIncNode(node.(addressableNode), false)
			continue
		}
		if p.consume("--") {
			node = newDecNode(node.(addressableNode), false)
			continue
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
				base, _, _ := p.baseType()
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
		init := newGVarInitStr(str)
		s := newGVar(newGVarLabel(), newTyArr(newTyChar(), utf8.RuneCountInString(str)), init)
		p.res.gVars = append(p.res.gVars, s)
		return newVarNode(s)
	}

	return newNumNode(int64(p.expectNum()))
}
