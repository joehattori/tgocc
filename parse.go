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
	res       *ast // TODO: maybe better to declare ast in global
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
	lvars, gvars, _ := p.curScope.segregateScopeVars()
	for _, v := range lvars {
		offset = alignTo(offset, v.getType().size())
		offset += v.getType().size()
		v.offset = offset + base
	}
	p.curScope.curOffset += offset
	p.curScope = p.curScope.super
	// maybe this is not necessary if we zero out the memory for child scope?
	p.curScope.curOffset += offset
	p.res.gVars = append(p.res.gVars, gvars...)
}

func (p *parser) findVar(s string) variable {
	v := p.searchVar(s)
	if v == nil {
		log.Fatalf("Undefined variable %s", s)
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

func (p *parser) consumeID() (tok *idTok, ok bool) {
	defer func() {
		if ok {
			p.popToks()
		}
	}()
	tok, ok = p.toks[0].(*idTok)
	return
}

func (p *parser) consumeStr() (tok *strTok, ok bool) {
	defer func() {
		if ok {
			p.popToks()
		}
	}()
	tok, ok = p.toks[0].(*strTok)
	return
}

func (p *parser) isEOF() (ok bool) {
	if len(p.toks) == 0 {
		return true
	}
	_, ok = p.toks[0].(*eofTok)
	return
}

func (p *parser) expect(str string) {
	if r, ok := p.toks[0].(*reservedTok); ok &&
		r.len == utf8.RuneCountInString(str) && strings.HasPrefix(r.str, str) {
		p.popToks()
		return
	}
	log.Fatalf("%s was expected but got %s", str, p.toks[0].getStr())
}

func (p *parser) expectID() (tok *idTok) {
	tok, _ = p.toks[0].(*idTok)
	if tok == nil {
		log.Fatalf("ID was expected but got %s", p.toks[0].getStr())
	}
	p.popToks()
	return
}

func (p *parser) expectNum() (tok *numTok) {
	tok, _ = p.toks[0].(*numTok)
	if tok == nil {
		log.Fatalf("Number was expected but got %s", p.toks[0].getStr())
	}
	p.popToks()
	return
}

func (p *parser) expectStr() (tok *strTok) {
	tok, _ = p.toks[0].(*strTok)
	if tok == nil {
		log.Fatalf("String literal was expected but got %s", p.toks[0].getStr())
	}
	p.popToks()
	return
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

func (p *parser) isType() (ret bool) {
	switch tok := p.toks[0].(type) {
	case *idTok:
		_, ret = p.searchVar(tok.name).(*typeDef)
	case *reservedTok:
		ret = typeMatcher.MatchString(tok.str)
	}
	return
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
				| "do" stmt "while" "(" expr ")" ";"
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
	for !p.isEOF() {
		if p.isFunction() {
			if fn := p.function(); fn != nil {
				ast.fns = append(ast.fns, fn)
			}
		} else {
			if ty, id, rhs, sc := p.decl(); ty != nil {
				init := buildGVarInit(ty, rhs)
				emit := (sc & extern) == 0
				p.curScope.addGVar(emit, id, ty, init)
				ast.gVars = append(ast.gVars, newGVar(emit, id, ty, init))
			}
		}
	}
}

func buildGVarInit(t ty, rhs node) gVarInit {
	if rhs == nil {
		return nil
	}
	switch t := t.(type) {
	case *tyArr:
		switch rhs := rhs.(type) {
		case *blkNode:
			var body []gVarInit
			idx := 0
			for _, e := range rhs.body {
				idx++
				body = append(body, buildGVarInit(t.of, e))
			}
			if t.len < 0 {
				t.len = idx
			}
			// zero out the rest
			if _, ok := t.of.(*tyArr); !ok {
				for i := idx; i < t.len; i++ {
					body = append(body, newGVarInitZero(t.of.size()))
				}
			}
			return newGVarInitArr(body)
		default:
			if ok, str := isStrNode(rhs); ok {
				if t.len < 0 {
					t.len = len(str)
				} else if t.len > len(str) {
					for i := len(str); i < t.len; i++ {
						str += string('\000')
					}
				}
				return newGVarInitStr(str)
			}
			log.Fatalf("Unhandled case in global variable initialization: %T", rhs)
			return nil
		}
	case *tyStruct:
		var body []gVarInit
		idx := 0
		for i, e := range rhs.(*blkNode).body {
			idx++
			mem := t.members[i]
			if e == nil {
				body = append(body, newGVarInitZero(mem.ty.size()))
				continue
			}
			var toAppend []gVarInit
			toAppend = append(toAppend, buildGVarInit(mem.ty, e))
			// padding for struct members
			var end int
			if i < len(t.members)-1 {
				end = t.members[i+1].offset
			} else {
				end = t.size()
			}
			start := mem.offset + mem.ty.size()
			if end-start > 0 {
				toAppend = append(toAppend, newGVarInitZero(end-start))
			}
			body = append(body, newGVarInitArr(toAppend))
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
	p.curScope.addGVar((sc&static) != 0, fnName, newTyFn(ty), nil)
	fn := newFnNode((sc&static) != 0, fnName, ty)
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

type storageClass int

const (
	static = 0b01
	extern = 0b10
)

func (p *parser) decl() (t ty, id string, rhs node, sc storageClass) {
	t, isTypeDef, sc := p.baseType()
	if p.consume(";") {
		return
	}
	id, t = p.tyDecl(t)
	if isTypeDef {
		p.expect(";")
		p.curScope.addTypeDef(id, t)
		// returned t is nil when it is typedef (no need to add to scope.vars)
		return nil, "", nil, sc
	}
	t = p.tySuffix(t)
	if p.consume(";") {
		return
	}
	if (sc & extern) != 0 {
		p.expect(";")
		return
	}
	p.expect("=")
	rhs = p.initializer(t, sc)
	p.expect(";")
	return
}

func (p *parser) initializer(t ty, sc storageClass) node {
	switch t := t.(type) {
	case *tyArr:
		var nodes []node
		if strTok, ok := p.consumeStr(); ok {
			init := newGVarInitStr(strTok.getStr())
			s := newGVar((sc&static) != 0, newGVarLabel(), newTyArr(newTyChar(), strTok.len), init)
			p.res.gVars = append(p.res.gVars, s)
			return newVarNode(s)
		}
		p.expect("{")
		for !p.consume("}") {
			nodes = append(nodes, p.initializer(t.of, sc))
			if !p.consume(",") {
				p.expect("}")
				break
			}
		}
		return newBlkNode(nodes)
	case *tyStruct:
		nodes := make([]node, len(t.members))
		if strTok, ok := p.consumeStr(); ok {
			init := newGVarInitStr(strTok.getStr())
			s := newGVar((sc&static) != 0, newGVarLabel(), newTyArr(newTyChar(), strTok.len), init)
			p.res.gVars = append(p.res.gVars, s)
			return newVarNode(s)
		}
		p.expect("{")
		idx := 0
		for !p.consume("}") {
			if p.consume(".") {
				id := p.expectID().name
				p.expect("=")
				for i, mem := range t.members[idx:] {
					if id == mem.name {
						idx += i
						break
					}
				}
			}
			nodes[idx] = p.initializer(t.members[idx].ty, sc)
			if !p.consume(",") {
				p.expect("}")
				break
			}
			idx++
		}
		return newBlkNode(nodes)
	default:
		return p.assign()
	}
}

func (p *parser) baseType() (t ty, isTypeDef bool, sc storageClass) {
	if p.consume("typedef") {
		isTypeDef = true
	}
	if p.consume("static") {
		sc |= static
	}
	if p.consume("extern") {
		sc |= extern
	}
	if isTypeDef && (sc != 0) || (sc == 0b11) {
		log.Fatal("typedef, static and extern should not be used together.")
	}
	cur := p.toks[0]
	switch tok := cur.(type) {
	case *idTok:
		id := idMatcher.FindString(tok.name)
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
			return newTyInt(), isTypeDef, sc
		}
		if p.consume("char") {
			return newTyChar(), isTypeDef, sc
		}
		if p.consume("short") {
			p.consume("int")
			return newTyShort(), isTypeDef, sc
		}
		if p.consume("long") {
			p.consume("long")
			p.consume("int")
			return newTyLong(), isTypeDef, sc
		}
		if p.consume("void") {
			return newTyVoid(), isTypeDef, sc
		}
		if p.consume("_Bool") {
			return newTyBool(), isTypeDef, sc
		}
	}
	log.Fatalf("Type expected but got %T: %s", cur, cur.getStr())
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
	return p.expectID().getStr(), p.tySuffix(baseTy)
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

func (p *parser) readFnParams(fn *fnNode) {
	p.expect("(")
	isFirstArg := true
	orig := p.toks
	if p.consume("void") && p.consume(")") {
		return
	}
	p.toks = orig
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
	tag, tagExists := p.consumeID()
	if tagExists && !p.beginsWith("{") {
		if tag := p.searchStructTag(tag.getStr()); tag != nil {
			return tag.ty
		}
		log.Fatalf("No such struct tag %s", tag.getStr())
	}
	p.expect("{")
	var members []*member
	offset, align := 0, 0
	for !p.consume("}") {
		// TODO: handle when rhs is not null
		ty, tag, _, _ := p.decl()
		offset = alignTo(offset, ty.alignment())
		members = append(members, newMember(tag, offset, ty))
		offset += ty.size()
		if align < ty.size() {
			align = ty.size()
		}
	}
	tyStruct := newTyStruct(align, members, alignTo(offset, align))
	if tagExists {
		p.curScope.addStructTag(newStructTag(tag.getStr(), tyStruct))
	}
	return tyStruct
}

func (p *parser) enumDecl() ty {
	p.expect("enum")
	tag, tagExists := p.consumeID()
	if tagExists && !p.beginsWith("{") {
		if tag := p.searchEnumTag(tag.getStr()); tag != nil {
			return tag.ty
		}
		log.Fatalf("No such enum tag %s", tag.getStr())
	}
	t := newTyEnum()

	p.expect("{")
	c := 0
	for {
		id := p.expectID()
		if p.consume("=") {
			c = int(p.constExpr())
		}
		p.curScope.addEnum(id.getStr(), t, c)
		c++
		orig := p.toks
		if p.consume("}") || p.consume(",") && p.consume("}") {
			break
		}
		p.toks = orig
		p.expect(",")
	}
	if tagExists {
		p.curScope.addEnumTag(newEnumTag(tag.getStr(), t))
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
		cond := p.expr()
		p.expect(")")
		then := p.stmt()
		return newWhileNode(cond, then)
	}

	// handle do-while statement
	if p.consume("do") {
		then := p.stmt()
		p.expect("while")
		p.expect("(")
		cond := p.expr()
		p.expect(")")
		p.expect(";")
		return newDoWhileNode(cond, then)
	}

	// handle for statement
	if p.consume("for") {
		p.expect("(")

		var init, cond, inc, then node

		p.spawnScope()
		if !p.consume(";") {
			if p.isType() {
				t, id, rhs, _ := p.decl()
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
		for idx := 0; ; idx++ {
			if c := p.switchCase(idx); c == nil {
				break
			} else {
				switch n := c.(type) {
				case *caseNode:
					cases = append(cases, n)
				case *defaultNode:
					if dflt != nil {
						log.Fatal("Multiple definition of default clause.")
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
		t, id, rhs, sc := p.decl()
		if id == "" {
			return newNullNode()
		}
		if (sc & static) != 0 {
			init := buildGVarInit(t, rhs)
			p.curScope.addGVar(true, id, t, init)
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
		isString, str := isStrNode(rhs)
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
		for i, mem := range rhs.(*blkNode).body {
			idx++
			node := newMemberNode(dst, t.members[i])
			if mem == nil {
				body = append(body, newExprNode(newAssignNode(node, newNumNode(0))))
			} else {
				body = append(body, newExprNode(newAssignNode(node, mem)))
			}
		}
		return newBlkNode(body)
	default:
		return newExprNode(newAssignNode(dst, rhs))
	}
}

func isStrNode(n node) (bool, string) {
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
			node = newBinaryNode(ndPtrAddEq, node.(addressableNode), p.assign())
		} else {
			node = newBinaryNode(ndAddEq, node.(addressableNode), p.assign())
		}
	} else if p.consume("-=") {
		if _, ok := node.loadType().(ptr); ok {
			node = newBinaryNode(ndPtrSubEq, node.(addressableNode), p.assign())
		} else {
			node = newBinaryNode(ndSubEq, node.(addressableNode), p.assign())
		}
	} else if p.consume("*=") {
		node = newBinaryNode(ndMulEq, node.(addressableNode), p.assign())
	} else if p.consume("/=") {
		node = newBinaryNode(ndDivEq, node.(addressableNode), p.assign())
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
		node = newBinaryNode(ndLogOr, node, p.logAnd())
	}
	return node
}

func (p *parser) logAnd() node {
	node := p.bitOr()
	for p.consume("&&") {
		node = newBinaryNode(ndLogAnd, node, p.bitOr())
	}
	return node
}

func (p *parser) bitOr() node {
	node := p.bitXor()
	for p.consume("|") {
		node = newBinaryNode(ndBitOr, node, p.bitXor())
	}
	return node
}

func (p *parser) bitXor() node {
	node := p.bitAnd()
	for p.consume("^") {
		node = newBinaryNode(ndBitXor, node, p.bitXor())
	}
	return node
}

func (p *parser) bitAnd() node {
	node := p.equality()
	for p.consume("&") {
		node = newBinaryNode(ndBitAnd, node, p.equality())
	}
	return node
}

func (p *parser) equality() node {
	node := p.relational()
	for {
		if p.consume("==") {
			node = newBinaryNode(ndEq, node, p.relational())
		} else if p.consume("!=") {
			node = newBinaryNode(ndNeq, node, p.relational())
		} else {
			return node
		}
	}
}

func (p *parser) relational() node {
	node := p.shift()
	for {
		if p.consume("<=") {
			node = newBinaryNode(ndLeq, node, p.shift())
		} else if p.consume(">=") {
			node = newBinaryNode(ndGeq, node, p.shift())
		} else if p.consume("<") {
			node = newBinaryNode(ndLt, node, p.shift())
		} else if p.consume(">") {
			node = newBinaryNode(ndGt, node, p.shift())
		} else {
			return node
		}
	}
}

func (p *parser) shift() node {
	node := p.addSub()
	for {
		if p.consume("<<") {
			node = newBinaryNode(ndShl, node, p.shift())
		} else if p.consume(">>") {
			node = newBinaryNode(ndShr, node, p.shift())
		} else if p.consume("<<=") {
			node = newBinaryNode(ndShlEq, node, p.shift())
		} else if p.consume(">>=") {
			node = newBinaryNode(ndShrEq, node, p.shift())
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
			node = newBinaryNode(ndMul, node, p.cast())
		} else if p.consume("/") {
			node = newBinaryNode(ndDiv, node, p.cast())
		} else {
			return node
		}
	}
}

func (p *parser) cast() node {
	orig := p.toks
	if p.consume("(") {
		if p.isType() {
			t, _, _ := p.baseType()
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
				mem := s.findMember(p.expectID().getStr())
				node = newMemberNode(node.(addressableNode), mem)
				continue
			}
			log.Fatalf("Expected struct but got %T", node.loadType())
		}
		if p.consume("->") {
			if t, ok := node.loadType().(*tyPtr); ok {
				mem := t.to.(*tyStruct).findMember(p.expectID().getStr())
				node = newMemberNode(newDerefNode(node.(addressableNode)), mem)
				continue
			}
			log.Fatalf("Expected pointer but got %T", node.loadType())
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
		log.Fatal("Statement expression returning void is not supported")
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
		id := id.getStr()
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
			log.Fatalf("Unhandled case of variable in primary: %T", p.findVar(id))
		}
	}

	if strTok, isStr := p.consumeStr(); isStr {
		init := newGVarInitStr(strTok.getStr())
		s := newGVar(true, newGVarLabel(), newTyArr(newTyChar(), strTok.len), init)
		p.res.gVars = append(p.res.gVars, s)
		return newVarNode(s)
	}

	return newNumNode(int64(p.expectNum().val))
}
