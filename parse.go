package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"
)

func (t *tokenized) popToks() {
	t.toks = t.toks[1:]
}

func (t *tokenized) consume(str string) bool {
	if r, ok := t.toks[0].(*reservedTok); ok &&
		r.len == len(str) &&
		strings.HasPrefix(r.str, str) {
		t.popToks()
		return true
	}
	return false
}

func (t *tokenized) consumeID() (string, bool) {
	cur := t.toks[0]
	id := idRegexp.FindString(cur.getStr())
	if _, ok := cur.(*idTok); ok && id != "" {
		t.popToks()
		return id, true
	}
	return "", false
}

func (t *tokenized) consumeStr() (string, bool) {
	if s, ok := t.toks[0].(*strTok); ok {
		t.popToks()
		return s.content, true
	}
	return "", false
}

func (t *tokenized) expect(str string) {
	cur := t.toks[0]
	if r, ok := cur.(*reservedTok); ok &&
		r.len == len(str) && strings.HasPrefix(cur.getStr(), str) {
		t.popToks()
		return
	}
	log.Fatalf("%s was expected but got %s", str, cur.getStr())
}

func (t *tokenized) expectID() string {
	cur := t.toks[0]
	id := idRegexp.FindString(cur.getStr())
	if _, ok := cur.(*idTok); ok && id != "" {
		t.popToks()
		return id
	}
	log.Fatalf("ID was expected but got %s", cur.getStr())
	return ""
}

func (t *tokenized) expectNum() int {
	cur := t.toks[0]
	if n, ok := cur.(*numTok); ok {
		t.popToks()
		return n.val
	}
	log.Fatalf("Number was expected but got %s", cur.getStr())
	return -1
}

func (t *tokenized) expectStructDecl() ty {
	t.expect("struct")
	tagStr, tagExists := t.consumeID()
	if tagExists && !t.beginsWith("{") {
		if tag := t.searchStructTag(tagStr); tag != nil {
			return tag.ty
		}
		log.Fatalf("no such struct tag %s", tagStr)
	}
	t.expect("{")
	var members []*member
	offset := 0
	for !t.consume("}") {
		// TODO: handle when rhs is not null
		ty, tag, _ := t.decl()
		members = append(members, newMember(tag, offset, ty))
		offset += ty.size()
	}
	tyStruct := newTyStruct(members, offset)
	if tagExists {
		t.curScope.addStructTag(newStructTag(tagStr, tyStruct))
	}
	return tyStruct
}

func (t *tokenized) expectType() ty {
	cur := t.toks[0]
	rs, ok := cur.(*reservedTok)
	if !ok {
		log.Fatalf("tkReserved was expected but got %d %s", cur, cur.getStr())
	}
	if strings.HasPrefix(rs.str, "int") {
		t.popToks()
		return newTyInt()
	}
	if strings.HasPrefix(rs.str, "char") {
		t.popToks()
		return newTyChar()
	}
	if strings.HasPrefix(rs.str, "struct") {
		return t.expectStructDecl()
	}
	return nil
}

func (t *tokenized) beginsWith(s string) bool {
	return strings.HasPrefix(t.toks[0].getStr(), s)
}

func (t *tokenized) isFunction() bool {
	orig := t.toks
	ty := t.expectType()
	for t.consume("*") {
		ty = newTyPtr(ty)
	}
	_, isID := t.consumeID()
	ret := isID && t.consume("(")
	t.toks = orig
	return ret
}

func (t *tokenized) isType() bool {
	r := regexp.MustCompile(`^(int|char|struct)\W`)
	return r.MatchString(t.toks[0].getStr())
}

var gVarLabelCount int

func newGVarLabel() string {
	ret := fmt.Sprintf(".L.data.%d", gVarLabelCount)
	gVarLabelCount++
	return ret
}

// program    = (function | globalVar)*
// function   = ty ident ("(" (ident ("," ident)* )? ")") "{" stmt* "}"
// globalVar  = decl
// stmt       = expr ";"
//				| "{" stmt* "}"
//			    | "return" expr ";"
//				| "if" "(" expr ")" stmt ("else" stmt) ?
//				| "while" "(" expr ")" stmt
//				| "for" "(" expr? ";" expr? ";" expr? ")" stmt
//				| decl
// decl       = ty ident ("[" expr "]")* "=" expr ;" | ty ";"
// expr       = assign
// stmtExpr   = "(" "{" stmt+ "}" ")"
// assign     = equality ("=" assign) ?
// equality   = relational ("==" relational | "!=" relational)*
// relational = add ("<" add | "<=" add | ">" add | ">=" add)*
// add        = mul ("+" mul | "-" mul)*
// mul        = unary ("*" unary | "/" unary)*
// unary      = "sizeof" unary | ("+" | "-" | "*" | "&")? unary | postfix
// postfix    = primary (("[" expr "]") | ("." | id))*
// primary    = num | str | ident ("(" (expr ("," expr)* )? ")")? | "(" expr ")" | stmtExpr

func (t *tokenized) readVarSuffix(ty ty) ty {
	if t.consume("[") {
		l := t.expectNum()
		t.expect("]")
		ty = t.readVarSuffix(ty)
		return newTyArr(ty, l)
	}
	return ty
}

func (t *tokenized) decl() (ty ty, id string, rhs node) {
	ty = t.expectType()
	for t.consume("*") {
		ty = newTyPtr(ty)
	}
	if t.consume(";") {
		return
	}
	id = t.expectID()
	ty = t.readVarSuffix(ty)
	if t.consume(";") {
		return
	}
	t.expect("=")
	rhs = t.expr()
	t.expect(";")
	return
}

func (t *tokenized) parse() {
	ast := t.res
	for {
		if _, ok := t.toks[0].(*eofTok); ok {
			break
		}
		if t.isFunction() {
			ast.fns = append(ast.fns, t.function())
		} else {
			// TODO: gvar init
			ty, id, _ := t.decl()
			g := newGVar(id, ty, nil)
			ast.gVars = append(ast.gVars, g)
			t.curScope.vars = append(t.curScope.vars, g)
		}
	}
}

func (t *tokenized) function() *fnNode {
	ty := t.expectType()
	for t.consume("*") {
		ty = newTyPtr(ty)
	}
	fnName := t.expectID()
	t.curFnName = fnName
	fn := newFnNode(fnName, ty)
	t.spawnScope()
	t.readFnParams(fn)
	t.expect("{")
	for !t.consume("}") {
		fn.body = append(fn.body, t.stmt())
	}
	t.setFnLVars(fn)
	t.rewindScope()
	// TODO: align
	fn.stackSize = t.curScope.curOffset
	return fn
}

func (t *tokenized) readFnParams(fn *fnNode) {
	t.expect("(")
	isFirstArg := true
	for !t.consume(")") {
		if !isFirstArg {
			t.expect(",")
		}
		isFirstArg = false

		ty := t.expectType()
		for t.consume("*") {
			ty = newTyPtr(ty)
		}
		lv := t.curScope.addLVar(t.expectID(), ty)
		fn.params = append(fn.params, lv)
	}
}

func (t *tokenized) setFnLVars(fn *fnNode) {
	offset := 0
	for _, sv := range t.curScope.vars {
		switch v := sv.(type) {
		case *lVar:
			offset += v.getType().size()
			v.offset = offset
			fn.lVars = append(fn.lVars, v)
		}
	}
}

func (t *tokenized) stmt() node {
	// handle block
	if t.consume("{") {
		var blkStmts []node
		t.spawnScope()
		for !t.consume("}") {
			blkStmts = append(blkStmts, t.stmt())
		}
		t.rewindScope()
		return newBlkNode(blkStmts)
	}

	// handle return
	if t.consume("return") {
		node := newRetNode(t.expr(), t.curFnName)
		t.expect(";")
		return node
	}

	// handle if statement
	if t.consume("if") {
		t.expect("(")
		cond := t.expr()
		t.expect(")")
		then := t.stmt()

		var els node
		if t.consume("else") {
			els = t.stmt()
		}

		return newIfNode(cond, then, els)
	}

	// handle while statement
	if t.consume("while") {
		t.expect("(")
		condNode := t.expr()
		t.expect(")")
		thenNode := t.stmt()
		return newWhileNode(condNode, thenNode)
	}

	// handle for statement
	if t.consume("for") {
		t.expect("(")

		var init, cond, inc, then node

		if !t.consume(";") {
			init = newExprNode(t.expr())
			t.expect(";")
		}

		if !t.consume(";") {
			cond = t.expr()
			t.expect(";")
		}

		if !t.consume(")") {
			inc = newExprNode(t.expr())
			t.expect(")")
		}

		then = t.stmt()
		return newForNode(init, cond, inc, then)
	}

	// handle variable definition
	if t.isType() {
		ty, id, rhs := t.decl()
		if id == "" {
			return newNullNode()
		}
		t.curScope.addLVar(id, ty)
		if rhs == nil {
			return newNullNode()
		}
		node := newAssignNode(newVarNode(t.findVar(id)), rhs)
		return newExprNode(node)
	}

	node := t.expr()
	t.expect(";")
	return newExprNode(node)
}

func (t *tokenized) expr() node {
	return t.assign()
}

func (t *tokenized) assign() node {
	node := t.equality()
	if t.consume("=") {
		assignNode := t.assign()
		node = newAssignNode(node.(addressableNode), assignNode)
	}
	return node
}

func (t *tokenized) equality() node {
	node := t.relational()
	for {
		if t.consume("==") {
			node = newArithNode(ndEq, node, t.relational())
		} else if t.consume("!=") {
			node = newArithNode(ndNeq, node, t.relational())
		} else {
			return node
		}
	}
}

func (t *tokenized) relational() node {
	node := t.addSub()
	for {
		if t.consume("<=") {
			node = newArithNode(ndLeq, node, t.addSub())
		} else if t.consume(">=") {
			node = newArithNode(ndGeq, node, t.addSub())
		} else if t.consume("<") {
			node = newArithNode(ndLt, node, t.addSub())
		} else if t.consume(">") {
			node = newArithNode(ndGt, node, t.addSub())
		} else {
			return node
		}
	}
}

func (t *tokenized) addSub() node {
	node := t.mulDiv()
	for {
		if t.consume("+") {
			node = newAddNode(node, t.mulDiv())
		} else if t.consume("-") {
			node = newSubNode(node, t.mulDiv())
		} else {
			return node
		}
	}
}

func (t *tokenized) mulDiv() node {
	node := t.unary()
	for {
		if t.consume("*") {
			node = newArithNode(ndMul, node, t.unary())
		} else if t.consume("/") {
			node = newArithNode(ndDiv, node, t.unary())
		} else {
			return node
		}
	}
}

func (t *tokenized) unary() node {
	if t.consume("sizeof") {
		return newNumNode(t.unary().loadType().size())
	}
	if t.consume("+") {
		return t.unary()
	}
	if t.consume("-") {
		node := t.unary()
		return newSubNode(newNumNode(0), node)
	}
	if t.consume("*") {
		node := t.unary()
		return newDerefNode(node)
	}
	if t.consume("&") {
		node := t.unary()
		return newAddrNode(node.(addressableNode))
	}
	return t.postfix()
}

func (t *tokenized) postfix() node {
	node := t.primary()
	for {
		if t.consume("[") {
			add := newAddNode(node, t.expr())
			node = newDerefNode(add)
			t.expect("]")
			continue
		}
		if t.consume(".") {
			if s, ok := node.loadType().(*tyStruct); ok {
				mem := s.findMember(t.expectID())
				node = newMemberNode(node.(addressableNode), mem)
				continue
			}
			log.Fatalf("expected struct but got %T", node.loadType())
		}
		if t.consume("->") {
			if p, ok := node.loadType().(*tyPtr); ok {
				mem := p.to.(*tyStruct).findMember(t.expectID())
				node = newMemberNode(newDerefNode(node.(addressableNode)), mem)
				continue
			}
			log.Fatalf("expected struct but got %T", node.loadType())
		}
		return node
	}
}

func (t *tokenized) stmtExpr() node {
	// "(" and "{" is already read.
	t.spawnScope()
	body := make([]node, 0)
	body = append(body, t.stmt())
	for !t.consume("}") {
		body = append(body, t.stmt())
	}
	t.expect(")")
	if ex, ok := body[len(body)-1].(*exprNode); !ok {
		log.Fatal("statement expression returning void is not supported")
	} else {
		body[len(body)-1] = ex.body
	}
	t.rewindScope()
	return newStmtExprNode(body)
}

func (t *tokenized) primary() node {
	if t.consume("(") {
		if t.consume("{") {
			return t.stmtExpr()
		}
		node := t.expr()
		t.expect(")")
		return node
	}

	if id, isID := t.consumeID(); isID {
		if t.consume("(") {
			var params []node
			if t.consume(")") {
				return newFnCallNode(id, params)
			}
			params = append(params, t.expr())
			for t.consume(",") {
				params = append(params, t.expr())
			}
			t.expect(")")
			return newFnCallNode(id, params)
		}
		return newVarNode(t.findVar(id))
	}

	if str, isStr := t.consumeStr(); isStr {
		s := newGVar(newGVarLabel(), newTyArr(newTyChar(), utf8.RuneCountInString(str)), str)
		t.res.gVars = append(t.res.gVars, s)
		return newVarNode(s)
	}

	return newNumNode(t.expectNum())
}
